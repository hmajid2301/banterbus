{
  description = "Development environment for BanterBus";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    pre-commit-hooks.url = "github:cachix/pre-commit-hooks.nix";
    playwright.url = "github:pietdevries94/playwright-web-flake/1.47.2";

    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    gomod2nix,
    pre-commit-hooks,
    playwright,
    ...
  }: (
    flake-utils.lib.eachDefaultSystem
    (system: let
      overlay = final: prev: {
        inherit (playwright.packages.${system}) playwright-test playwright-driver;
      };
      pkgs = import nixpkgs {
        inherit system;
        overlays = [overlay];
      };

      myPackages = with pkgs; [
        go_1_22
        playwright-test

        goose
        air
        golangci-lint
        gotools
        gotestsum
        gocover-cobertura
        go-task
        go-mockery
        goreleaser
        golines

        tailwindcss
        templ
        sqlc
        gitlab-ci-local
      ];

      # The current default sdk for macOS fails to compile go projects, so we use a newer one for now.
      # This has no effect on other platforms.
      callPackage = pkgs.darwin.apple_sdk_11_0.callPackage or pkgs.callPackage;
    in rec {
      packages.default = callPackage ./. {
        inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
      };
      devShells.default = callPackage ./shell.nix {
        inherit (gomod2nix.legacyPackages.${system}) mkGoEnv gomod2nix;
        inherit pre-commit-hooks;
        inherit myPackages;
      };
      packages.container = pkgs.callPackage ./containers/service.nix {package = packages.default;};
      packages.container-ci = pkgs.callPackage ./containers/ci.nix {
        inherit (gomod2nix.legacyPackages.${system}) mkGoEnv gomod2nix;
        inherit myPackages;
      };
    })
  );
}
