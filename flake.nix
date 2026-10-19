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

      # Evaluates the NixOS module in a minimal system, including its
      # assertions, so a broken module fails `nix flake check`. Linux only.
      checks = nixpkgs.lib.genAttrs [ "x86_64-linux" "aarch64-linux" ] (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          evaluated = nixpkgs.lib.nixosSystem {
            inherit system;
            modules = [
              self.nixosModules.default
              {
                services.ping = {
                  enable = true;
                  hosts = [ "example.com" ];
                };
                boot.isContainer = true;
                system.stateVersion = "26.05";
              }
            ];
          };
          assertionsHold = builtins.all (a: a.assertion) evaluated.config.assertions;
        in
        {
          nixos-module =
            assert assertionsHold;
            pkgs.writeText "ping-nixos-module-unit" (
              builtins.toJSON evaluated.config.systemd.services.ping.serviceConfig
            );
        }
      );

      formatter = forAllSystems (pkgs: pkgs.nixfmt-rfc-style);
    };
}
