{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        packages.default = pkgs.buildGoModule {
          pname = "chirpy";
          version = "v0.0.0";
          src = ./.;
          vendorHash = null;
        };
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls

            postgresql
            goose
            sqlc
          ];
        };

        formatter = pkgs.alejandra;
      }
    );
}
