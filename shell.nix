{
  pkgs ? (
    let
      inherit (builtins) fetchTree fromJSON readFile;
      inherit ((fromJSON (readFile ./flake.lock)).nodes) nixpkgs gomod2nix;
    in
      import (fetchTree nixpkgs.locked) {
        overlays = [
          (import "${fetchTree gomod2nix.locked}/overlay.nix")
        ];
      }
  ),
  mkGoEnv ? pkgs.mkGoEnv,
  gomod2nix ? pkgs.gomod2nix,
  pre-commit-hooks,
  ...
}: let
  goEnv = mkGoEnv {pwd = ./.;};
  pre-commit-check = pre-commit-hooks.lib.${pkgs.system}.run {
    src = ./.;
    hooks = {
      golangci-lint.enable = true;
      gotest.enable = true;
    };
  };
in
  pkgs.mkShell {
    inherit (pre-commit-check) shellHook;
    hardeningDisable = ["all"];
    packages = with pkgs; [
      # TODO: workout how to use go env
      # goEnv
      gomod2nix
      go_1_22

      tailwindcss
      docker

      air
      golangci-lint
      gotools
      gotestsum
      go-junit-report
      gocover-cobertura
      go-task
      go-mockery
      goreleaser
      golines
      templ

      sqlc
    ];
  }
