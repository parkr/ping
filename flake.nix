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

      formatter = forAllSystems (pkgs: pkgs.nixfmt-rfc-style);
    };
}
