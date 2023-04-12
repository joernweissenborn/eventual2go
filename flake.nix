{
  inputs = {
    nixpkgs = {
      url = "github:nixos/nixpkgs/nixos-unstable";
    };
    flake-utils = {
      url = "github:numtide/flake-utils";
    };
  };
  outputs = { nixpkgs, flake-utils, ... }: flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        inherit system;
      };
      python = pkgs.python310;
      lib-path = with pkgs; lib.makeLibraryPath [
        stdenv.cc.cc
      ];


    in
    rec {
      devShell = pkgs.mkShell {
        buildInputs = with pkgs; [
          go_1_20
          pre-commit
        ];
        shellHook = ''
          pre-commit install
        '';

      };
    }
  );
}
