{
  pkgs,
  mkGoEnv,
  gomod2nix,
  pre-commit-hooks,
  ...
}:
# INFO: this creates an image from the dev shell, but doesn't work with gitlab.
# So we are now using the image below for now.
# pkgs.dockerTools.buildNixShellImage {
#   name = "banterbus-dev";
#   tag = "latest";
#   drv = (import ./shell.nix) {
#     inherit pkgs mkGoEnv gomod2nix pre-commit-hooks;
#   };
#   shell = "${pkgs.bash}/bin/sh";
# }
pkgs.dockerTools.buildImage {
  name = "banterbus-dev";
  tag = "latest";
  copyToRoot = pkgs.buildEnv {
    name = "banterbus-dev";
    pathsToLink = ["/bin"];
    # TODO: pull from shell.nix
    paths = with pkgs; [
      coreutils
      gnugrep
      nix
      bash
      go_1_22
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
      dockerTools.caCertificates
      cacert.out
      which
      curl
      git
    ];
  };
  config = {
    Env = [
      "NIX_PAGER=cat"
      # A user is required by nix
      # https://github.com/NixOS/nix/blob/9348f9291e5d9e4ba3c4347ea1b235640f54fd79/src/libutil/util.cc#L478
      "USER=nobody"
      "SSL_CERT_FILE=${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt"
      "SSL_CERT_DIR=${pkgs.cacert}/etc/ssl/certs/"
    ];
  };
}
