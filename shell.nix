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
    hardeningDisable = ["all"];
    shellHook = ''
      export PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1
      export PLAYWRIGHT_NODEJS_PATH="${pkgs.nodejs}/bin/node"
      export PLAYWRIGHT_BROWSERS_PATH="${pkgs.playwright-driver.browsers}"
      export PLAYWRIGHT_PATH="${pkgs.playwright-test}/lib/node_modules/@playwright/test/cli.js"
      ${pre-commit-check.shellHook}
    '';
    buildInputs = pre-commit-check.enabledPackages;
    packages = with pkgs; [
      # TODO: workout how to use go env
      # goEnv
      gomod2nix
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
    ];
  }
