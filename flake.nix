{
  description = "ping: a small analytics ping server";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (pkgs: rec {
        ping = pkgs.callPackage ./nix/package.nix {
          rev = self.shortRev or self.dirtyShortRev or "dev";
        };
        default = ping;
      });

      # `import` this in your own config for `pkgs.ping`:
      #   nixpkgs.overlays = [ inputs.ping.overlays.default ];
      overlays.default = final: _prev: {
        ping = final.callPackage ./nix/package.nix {
          rev = self.shortRev or self.dirtyShortRev or "dev";
        };
      };

      # `services.ping.enable = true;` -- see nix/module.nix for the options.
      #   imports = [ inputs.ping.nixosModules.default ];
      nixosModules.default = import ./nix/module.nix self;

      apps = forAllSystems (pkgs: {
        default = {
          type = "app";
          program = "${self.packages.${pkgs.system}.ping}/bin/ping";
        };
      });

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = [
            pkgs.go
            pkgs.gopls
          ];
        };
      });

      # Tests the NixOS module without booting a VM: evaluates it in minimal
      # systems and asserts on the result (including that bad input is rejected),
      # then runs the real binary with the exact arguments the module generates.
      # Linux only.
      checks = nixpkgs.lib.genAttrs [ "x86_64-linux" "aarch64-linux" ] (
        system:
        let
          inherit (nixpkgs) lib;
          pkgs = nixpkgs.legacyPackages.${system};

          # Just enough for a system to evaluate: no bootloader, a stub root.
          fixture = {
            boot.loader.grub.enable = false;
            fileSystems."/" = {
              device = "none";
              fsType = "tmpfs";
            };
            system.stateVersion = "26.05";
          };

          # A system with services.ping enabled; `settings` overrides the
          # defaults below, `extra` is any other configuration.
          evalWith =
            settings: extra:
            lib.nixosSystem {
              inherit system;
              modules = [
                self.nixosModules.default
                fixture
                extra
                {
                  services.ping = {
                    enable = true;
                    hosts = [ "example.com" ];
                  }
                  // settings;
                }
              ];
            };
          service = settings: (evalWith settings { }).config.systemd.services.ping;
          exec = settings: (service settings).serviceConfig.ExecStart;
          failedAssertions =
            settings: extra:
            map (a: a.message) (lib.filter (a: !a.assertion) (evalWith settings extra).config.assertions);
          # Option types reject bad values with an evaluation error.
          rejected =
            settings: !(builtins.tryEval (builtins.deepSeq (service settings).serviceConfig true)).success;
          has = lib.hasInfix;
          check = lib.assertMsg;

          tests = [
            (check (has ''"-http=127.0.0.1:8000"'' (exec { })) "listens on 127.0.0.1:8000 by default")
            (check (has ''"-baseurl=http://localhost:8000"'' (exec { })) "baseUrl defaults from the port")
            (check (has ''"-hosts=example.com"'' (exec { })) "hosts reach -hosts")
            (check (!((service { }).environment ? PORT)) "PORT is not set: -http is the only port input")
            (check (
              (service { }).environment.PING_DB == "/var/lib/ping/ping_production.sqlite3"
            ) "database lives in the state directory")
            (check (lib.hasPrefix "ping-"
              (evalWith { } { }).config.services.ping.package.name
            ) "package is built")

            # hosts: required, and each entry must be a single non-empty token.
            (check (failedAssertions { hosts = [ ]; } { } != [ ]) "empty hosts is rejected by an assertion")
            (check (rejected { hosts = [ "" ]; }) "an empty host is rejected")
            (check (rejected { hosts = [ "a.com,b.com" ]; }) "a host with a comma is rejected")
            (check (rejected { hosts = [ "a com" ]; }) "a host with whitespace is rejected")
            (check (has ''"-hosts=a.com,b.com:8443"'' (exec {
              hosts = [
                "a.com"
                "b.com:8443"
              ];
            })) "hosts join with commas and keep a port")

            # Other types.
            (check (rejected { port = 0; }) "port 0 is rejected")
            (check (rejected { port = 70000; }) "port 70000 is rejected")
            (check (rejected { baseUrl = "ftp://ping.example.com"; }) "a non-http baseUrl is rejected")
            (check (rejected { listenAddress = "a b"; }) "a listenAddress with whitespace is rejected")
            (check (rejected { stateDirectoryMode = "700"; }) "a 3-digit mode is rejected")
            (check (rejected { stateDirectoryMode = "0900"; }) "a non-octal mode is rejected")

            # IPv6 and systemd escaping.
            (check (has ''"-http=[::1]:8000"'' (exec {
              listenAddress = "::1";
            })) "a bare IPv6 address is bracketed")
            (check (has ''"-http=[::1]:8000"'' (exec {
              listenAddress = "[::1]";
            })) "an already-bracketed one is left alone")
            (check (has "%%eth0" (exec {
              listenAddress = "fe80::1%eth0";
            })) "a zone id's % is escaped for systemd")
            (check (
              let
                e = exec { baseUrl = "https://x/a%20b?q=$1"; };
              in
              has "%%20" e && has "$$1" e
            ) "% and $ in an argument are escaped for systemd")

            # Ports and capabilities.
            (check (!((service { }).serviceConfig ? AmbientCapabilities)) "no ambient capabilities by default")
            (check (
              (service { }).serviceConfig.CapabilityBoundingSet == ""
            ) "all capabilities dropped by default")
            (check (
              (service { port = 80; }).serviceConfig.AmbientCapabilities == [ "CAP_NET_BIND_SERVICE" ]
              && (service { port = 80; }).serviceConfig.CapabilityBoundingSet == [ "CAP_NET_BIND_SERVICE" ]
            ) "a port below 1024 gets CAP_NET_BIND_SERVICE")
            (check (
              (evalWith { openFirewall = true; } { }).config.networking.firewall.allowedTCPPorts == [ 8000 ]
              && (evalWith { } { }).config.networking.firewall.allowedTCPPorts == [ ]
            ) "openFirewall opens the port, and only when asked")

            # Users, groups and the state directory.
            (check ((evalWith { } { }).config.users.users ? ping) "the default user is created")
            (check (failedAssertions { user = "web"; } { } != [ ]) "an undefined custom user is rejected")
            (check (failedAssertions { group = "web"; } { } != [ ]) "an undefined custom group is rejected")
            (check (
              let
                custom = {
                  users.users.web = {
                    isSystemUser = true;
                    group = "web";
                  };
                  users.groups.web = { };
                };
              in
              failedAssertions {
                user = "web";
                group = "web";
              } custom == [ ]
              && !(
                (evalWith {
                  user = "web";
                  group = "web";
                } custom).config.users.users
                  ? ping
              )
            ) "a defined custom user and group are used, and ping's isn't created")
            (check (
              (service { }).serviceConfig.StateDirectoryMode == "0700"
              && (service { }).serviceConfig.UMask == "0027"
            ) "the state directory is private by default")
            (check (
              (service { stateDirectoryMode = "0750"; }).serviceConfig.StateDirectoryMode == "0750"
            ) "the state directory mode is configurable")
          ];

          # The real binary, started with the arguments and environment the
          # module generates for these settings, in the build sandbox (which
          # has loopback). Proves the flags parse, the allowlist works, and a
          # hit is recorded.
          runtimeSettings = {
            port = 18972;
            hosts = [ "example.com" ];
          };
          runtimeSystem = evalWith runtimeSettings { };
          # ExecStart is `"binary" "arg" ...`; none of these args need unescaping.
          unquote = a: lib.removeSuffix "\"" (lib.removePrefix "\"" a);
          runtimeArgs = map unquote (lib.tail (lib.splitString " " (exec runtimeSettings)));
          runtimeBinary = lib.getExe runtimeSystem.config.services.ping.package;
        in
        {
          nixos-module =
            assert builtins.all (t: t) tests;
            pkgs.runCommand "ping-nixos-module-tests" { } "touch $out";

          nixos-module-runtime =
            pkgs.runCommand "ping-nixos-module-runtime"
              {
                nativeBuildInputs = [ pkgs.curl ];
              }
              ''
                export PING_DB="$TMPDIR/ping.sqlite3"   # the module's path is /var/lib
                ${runtimeBinary} ${lib.escapeShellArgs runtimeArgs} &
                pid=$!
                trap 'kill $pid' EXIT
                url=http://127.0.0.1:${toString runtimeSettings.port}
                for _ in $(seq 1 50); do curl -sf $url/_health >/dev/null && break; sleep 0.2; done
                curl -sf $url/_health >/dev/null
                curl -sf -A check -H 'Referer: https://example.com/p' -o /dev/null $url/ping.js
                curl -sf "$url/counts?host=example.com&path=/p" | grep -q '"views":1'
                code=$(curl -s -o /dev/null -w '%{http_code}' -A check -H 'Referer: https://other.example/p' $url/ping.js)
                test "$code" = 401
                touch $out
              '';
        }
      );

      formatter = forAllSystems (pkgs: pkgs.nixfmt-rfc-style);
    };
}
