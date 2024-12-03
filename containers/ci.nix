{
  pkgs,
  myPackages,
  mkGoEnv,
  gomod2nix,
  ...
}: let
  goEnv = mkGoEnv {pwd = ../.;};
in
  pkgs.dockerTools.buildImage {
    name = "banterbus-dev";
    tag = "latest";
    copyToRoot = pkgs.buildEnv {
      name = "banterbus-dev";
      pathsToLink = ["/bin"];
      paths = with pkgs;
        [
          coreutils
          gnugrep
          bash
          curl
          git
          goEnv
          gomod2nix
        ]
        ++ myPackages;
    };
    config = {
      Env = [
        "NIX_PAGER=cat"
        # A user is required by nix
        # https://github.com/NixOS/nix/blob/9348f9291e5d9e4ba3c4347ea1b235640f54fd79/src/libutil/util.cc#L478
        "USER=nobody"
        "SSL_CERT_FILE=${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt"
        "SSL_CERT_DIR=${pkgs.cacert}/etc/ssl/certs/"
        "PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1"
        "PLAYWRIGHT_BROWSERS_PATH=${pkgs.playwright-driver.browsers}"
        "PLAYWRIGHT_NODEJS_PATH=${pkgs.nodejs}/bin/node"
        "PLAYWRIGHT_DRIVER_PATH=${pkgs.playwright-driver}"
      ];
    };
  }
