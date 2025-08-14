{
  description = "Flat Go/Echo service dev shell for pinjol";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; };
      in {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            golangci-lint
            gotestsum
            air
          ];
          shellHook = ''
            echo "Go version: $(go version)"
            echo "Run: make test | make run"
          '';
        };
      });
}
