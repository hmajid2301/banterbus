{
  pkgs,
  myPackages,
  ...
}:
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
        nix
        bash
        dockerTools.caCertificates
        cacert.out
        which
        curl
        git
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
    ];
  };
}
