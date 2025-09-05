{
  description = "Development environment for BanterBus";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    pre-commit-hooks.url = "github:cachix/pre-commit-hooks.nix";
    playwright.url = "github:pietdevries94/playwright-web-flake/1.52.0";

    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      gomod2nix,
      pre-commit-hooks,
      playwright,
      ...
    }:
    (flake-utils.lib.eachDefaultSystem (
      system:
      let
        overlay = final: prev: {
          inherit (playwright.packages.${system})
            playwright-test
            playwright-driver
            ;
          go-enum = prev.callPackage ./nix/go-enum.nix { };
        };
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ overlay ];
        };

        myPackages = with pkgs; [
          go_1_24

          goose
          air
          golangci-lint
          gotools
          gotestsum
          gocover-cobertura
          go-task
          go-mockery
          go-enum
          goreleaser
          golines

          playwright-driver
          tailwindcss
          templ
          sqlc
          concurrently
          nodePackages.prettier

          nixpacks
          kamal

          sqlfluff
          rustywind
        ];

        devShellPackages =
          with pkgs;
          myPackages
          ++ [
            gitlab-ci-local
            gum
            attic-client
          ];
      in
      rec {
        packages.default = pkgs.callPackage ./. {
          inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
        };
        devShells.default = pkgs.callPackage ./shell.nix {
          inherit (gomod2nix.legacyPackages.${system}) mkGoEnv gomod2nix;
          inherit pre-commit-hooks;
          inherit devShellPackages;
        };
        packages.container = pkgs.callPackage ./containers/service.nix {
          package = packages.default;
        };
        packages.container-ci = pkgs.callPackage ./containers/ci.nix {
          inherit (gomod2nix.legacyPackages.${system}) mkGoEnv gomod2nix;
          inherit myPackages;
        };
      }
    ));
}
