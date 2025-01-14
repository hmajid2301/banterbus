{
  pkgs ? (let
    inherit (builtins) fetchTree fromJSON readFile;
    inherit ((fromJSON (readFile ./flake.lock)).nodes) nixpkgs gomod2nix;
  in
    import (fetchTree nixpkgs.locked) {
      overlays = [(import "${fetchTree gomod2nix.locked}/overlay.nix")];
    }),
  mkGoEnv ? pkgs.mkGoEnv,
  gomod2nix ? pkgs.gomod2nix,
  pre-commit-hooks,
  devShellPackages,
  ...
}: let
  goEnv = mkGoEnv {pwd = ./.;};
  pre-commit-check = pre-commit-hooks.lib.${pkgs.system}.run {
    src = ./.;
    hooks = {
      golangci-lint.enable = true;
      gotest.enable = true;
      golines = {
        enable = true;
        settings = {flags = "-m 120";};
      };
      generate = {
        enable = true;
        name = "Generate code such as tailwindcss, templ components and sqlc";
        entry = "task generate";
        pass_filenames = false;
      };
    };
  };
in
  pkgs.mkShell {
    hardeningDisable = ["all"];
    shellHook = ''
      export PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1
      export PLAYWRIGHT_BROWSERS_PATH="${pkgs.playwright-driver.browsers}"
      export PLAYWRIGHT_NODEJS_PATH="${pkgs.nodejs}/bin/node"
      export PLAYWRIGHT_DRIVER_PATH="${pkgs.playwright-driver}"
      export GOOSE_MIGRATION_DIR="./db/migrations"
      export TEMPL_EXPERIMENT=rawgo
      export BANTERBUS_LOG_LEVEL=debug
      export BANTERBUS_DB_HOST=localhost
      export BANTERBUS_DB_USERNAME=postgres
      export BANTERBUS_DB_PASSWORD=postgres
      export BANTERBUS_DB_NAME=postgres
      export BANTERBUS_REDIS_ADDRESS="localhost:6379"
      export BANTERBUS_PLAYWRIGHT_HEADLESS=false
      export BANTERBUS_ENVIRONMENT=local
      export BANTERBUS_AUTO_RECONNECT=true
      export BANTERBUS_JWKS_URL=http://localhost:9090/default/jwks
      export CGO_ENABLED=0
      export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317

      ${pre-commit-check.shellHook}
    '';
    buildInputs = pre-commit-check.enabledPackages;
    packages = devShellPackages ++ [goEnv gomod2nix pkgs.gitlab-ci-local];
  }
