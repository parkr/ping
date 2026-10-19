# NixOS module for ping: `services.ping.enable = true;` runs the server as a
# hardened systemd service with its sqlite database in /var/lib/ping.
#
# Imported as `inputs.ping.nixosModules.default`. It takes the flake itself so
# `package` can default to this flake's build, with no overlay to wire up.
self:
{
  config,
  lib,
  pkgs,
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
in
{
  options.services.ping = {
    enable = mkEnableOption "ping, a small analytics ping server";

    package = mkOption {
      type = types.package;
      default = self.packages.${pkgs.stdenv.hostPlatform.system}.default;
      defaultText = lib.literalExpression "inputs.ping.packages.\${system}.default";
      description = "The ping package to run.";
    };

    hosts = mkOption {
      type = types.listOf types.str;
      example = [
        "example.com"
        "my-site.com"
      ];
      description = ''
        Hostnames whose visits are recorded: a request is accepted only when
        its Referer's host is in this list (`-hosts`). Required, because an
        empty list accepts nothing.
      '';
    };

    listenAddress = mkOption {
      type = types.str;
      default = "127.0.0.1";
      description = "Address to listen on. Loopback by default, for a reverse proxy.";
    };

    port = mkOption {
      type = types.port;
      default = 8000;
      description = "Port to listen on.";
    };

    baseUrl = mkOption {
      type = types.str;
      default = "http://localhost:${toString cfg.port}";
      defaultText = lib.literalExpression ''"http://localhost:''${toString config.services.ping.port}"'';
      example = "https://ping.example.com";
      description = "Public URL that /ping.js and /stats.js call back to (`-baseurl`).";
    };

    databasePath = mkOption {
      type = types.path;
      default = "/var/lib/ping/ping_production.sqlite3";
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
    ];

    users.users.ping = {
      isSystemUser = true;
      group = "ping";
    };
    users.groups.ping = { };

    systemd.services.ping = {
      description = "ping, a little analytics service";
      after = [ "network.target" ];
      wantedBy = [ "multi-user.target" ];
      environment = {
        PING_DB = cfg.databasePath;
        PORT = toString cfg.port;
      };
      serviceConfig = {
        ExecStart = lib.escapeShellArgs [
          (lib.getExe cfg.package)
          "-http=${cfg.listenAddress}:${toString cfg.port}"
          "-hosts=${lib.concatStringsSep "," cfg.hosts}"
          "-baseurl=${cfg.baseUrl}"
        ];
        User = "ping";
        Group = "ping";
        StateDirectory = "ping";
        StateDirectoryMode = "0700";
        Restart = "on-failure";
        RestartSec = 5;

        # Parses untrusted internet input, so it gets only the sqlite file.
        NoNewPrivileges = true;
        CapabilityBoundingSet = "";
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
      };
    };
  };
}
