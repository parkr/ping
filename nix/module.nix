# NixOS module for ping: `services.ping.enable = true;` runs the server as a
# hardened systemd service with its sqlite database in /var/lib/ping.
#
# Imported as `inputs.ping.nixosModules.default`. It takes the flake itself only
# to label the package with its revision.
self:
{
  config,
  lib,
  pkgs,
  utils,
  ...
}:

let
  cfg = config.services.ping;
  inherit (lib)
    mkOption
    mkEnableOption
    mkIf
    types
    ;

  # systemd's StateDirectory is relative to /var/lib. The database path is
  # derived from it so the two can't drift.
  stateDirectory = "ping";

  # An IPv6 literal needs brackets in host:port; leave an already-bracketed one.
  bindHost =
    if lib.hasPrefix "[" cfg.listenAddress || !(lib.hasInfix ":" cfg.listenAddress) then
      cfg.listenAddress
    else
      "[${cfg.listenAddress}]";

  # Binding below 1024 needs CAP_NET_BIND_SERVICE, which the service otherwise
  # drops along with every other capability.
  privilegedPort = cfg.port < 1024;

  noSpaces = "[^[:space:]]+";
in
{
  options.services.ping = {
    enable = mkEnableOption "ping, a small analytics ping server";

    package = mkOption {
      type = types.package;
      default = pkgs.callPackage ./package.nix {
        rev = self.shortRev or self.dirtyShortRev or "dev";
      };
      defaultText = lib.literalExpression "(built from this flake's source with your nixpkgs)";
      description = ''
        The ping package to run. By default it is built from this flake's source
        with the *consumer's* `pkgs`, so it shares your nixpkgs (and its Go and
        libc) rather than pulling in a second copy.
      '';
    };

    hosts = mkOption {
      type = types.listOf (types.strMatching "[^,[:space:]]+");
      example = [
        "example.com"
        "my-site.com"
        "example.com:8443"
      ];
      description = ''
        Sites whose visits are recorded (`-hosts`). A request is accepted only
        when its Referer's host is listed, compared exactly as the browser sends
        it: **a site on a non-default port must include the port**
        (`example.com:8443`). Entries can't be empty or contain commas or
        whitespace. Required, because an empty list accepts nothing.
      '';
    };

    user = mkOption {
      type = types.nonEmptyStr;
      default = "ping";
      description = ''
        User to run as. Created (as a system user) only when left at the
        default; otherwise it must be defined elsewhere in your configuration.
      '';
    };

    group = mkOption {
      type = types.nonEmptyStr;
      default = "ping";
      description = ''
        Group to run as, and the group of the state directory. Created only when
        left at the default; otherwise it must be defined elsewhere.
      '';
    };

    listenAddress = mkOption {
      type = types.strMatching noSpaces;
      default = "127.0.0.1";
      example = "[::1]";
      description = ''
        Address to listen on (`-http`), loopback by default for a reverse proxy.
        A bare IPv6 address is bracketed for you.
      '';
    };

    port = mkOption {
      type = types.ints.between 1 65535;
      default = 8000;
      description = ''
        Port to listen on (`-http`). A port below 1024 gets
        `CAP_NET_BIND_SERVICE` automatically.

        Only the `-http` flag is set, never the `PORT` environment variable the
        binary also reads: when both exist the flag wins, and the only other
        reader of `PORT` is `ping-healthcheck`, which this module doesn't run.
      '';
    };

    baseUrl = mkOption {
      type = types.strMatching "https?://${noSpaces}";
      default = "http://localhost:${toString cfg.port}";
      defaultText = lib.literalExpression ''"http://localhost:''${toString config.services.ping.port}"'';
      example = "https://ping.example.com";
      description = ''
        Public URL that `/ping.js` and `/stats.js` call back to (`-baseurl`).
        Always passed explicitly, so the binary's own default is never used.
      '';
    };

    openFirewall = mkOption {
      type = types.bool;
      default = false;
      description = ''
        Open `port` in the firewall. Leave off behind a reverse proxy on the
        default loopback address; turn on only if you set a non-loopback
        `listenAddress` and want ping reachable directly.
      '';
    };

    stateDirectoryMode = mkOption {
      type = types.strMatching "[0-7]{4}";
      default = "0700";
      example = "0750";
      description = ''
        Mode of the state directory (`/var/lib/ping`), and so who can read the
        database. The default lets only the service user and root in; the
        database file is created group-readable (`UMask=0027`), so the
        directory is what gates access.

        To let a backup that runs as another user read it, set `"0750"` and run
        that backup as `group` (or add its user to the group with
        `users.users.<name>.extraGroups`), and open the database read-only:
        `sqlite3 -readonly ${config.services.ping.databasePath}`. A backup run
        by root, the systemd default, needs no change.
      '';
    };

    databasePath = mkOption {
      type = types.str;
      default = "/var/lib/${stateDirectory}/ping_production.sqlite3";
      readOnly = true;
      description = ''
        The sqlite database. Fixed under the service's state directory, which
        is the only place it can write; exposed so a backup can find it.
      '';
    };
  };

  config = mkIf cfg.enable {
    assertions = [
      {
        assertion = cfg.hosts != [ ];
        message = "services.ping.hosts must list at least one hostname: ping accepts no visits from hosts it isn't told about.";
      }
      {
        assertion = cfg.user == "ping" || config.users.users ? ${cfg.user};
        message = "services.ping.user is \"${cfg.user}\", which isn't defined. Define users.users.${cfg.user}, or leave it at the default to have the module create \"ping\".";
      }
      {
        assertion = cfg.group == "ping" || config.users.groups ? ${cfg.group};
        message = "services.ping.group is \"${cfg.group}\", which isn't defined. Define users.groups.${cfg.group}, or leave it at the default to have the module create \"ping\".";
      }
    ];

    users.users = mkIf (cfg.user == "ping") {
      ping = {
        isSystemUser = true;
        inherit (cfg) group;
      };
    };
    users.groups = mkIf (cfg.group == "ping") { ping = { }; };

    networking.firewall.allowedTCPPorts = mkIf cfg.openFirewall [ cfg.port ];

    systemd.services.ping = {
      description = "ping, a little analytics service";
      after = [ "network.target" ];
      wantedBy = [ "multi-user.target" ];
      environment.PING_DB = cfg.databasePath;
      serviceConfig = {
        # systemd expands `%` specifiers and `$VAR` even inside quotes, so the
        # arguments need its escaping, not the shell's.
        ExecStart = utils.escapeSystemdExecArgs [
          (lib.getExe cfg.package)
          "-http=${bindHost}:${toString cfg.port}"
          "-hosts=${lib.concatStringsSep "," cfg.hosts}"
          "-baseurl=${cfg.baseUrl}"
        ];
        User = cfg.user;
        Group = cfg.group;
        StateDirectory = stateDirectory;
        StateDirectoryMode = cfg.stateDirectoryMode;
        UMask = "0027";
        Restart = "on-failure";
        RestartSec = 5;

        # Parses untrusted internet input, so it gets only the sqlite file.
        NoNewPrivileges = true;
        CapabilityBoundingSet = if privilegedPort then [ "CAP_NET_BIND_SERVICE" ] else "";
        ProtectSystem = "strict";
        ProtectHome = true;
        PrivateTmp = true;
        PrivateDevices = true;
        ProtectKernelTunables = true;
        ProtectKernelModules = true;
        ProtectControlGroups = true;
        RestrictAddressFamilies = [
          "AF_INET"
          "AF_INET6"
          "AF_UNIX"
        ];
        RestrictNamespaces = true;
        LockPersonality = true;
      }
      // lib.optionalAttrs privilegedPort { AmbientCapabilities = [ "CAP_NET_BIND_SERVICE" ]; };
    };
  };
}
