{
  pkgs,
  package,
}:
pkgs.dockerTools.buildImage {
  name = "banterbus";
  tag = "latest";
  created = "now";
  copyToRoot = pkgs.buildEnv {
    name = "image-root";
    paths = [
      package
      pkgs.cacert
    ];
    pathsToLink = ["/bin"];
  };
  config = {
    ExposedPorts = {
      "8080/tcp" = {};
    };
    Cmd = ["${package}/bin/banterbus"];
    Env = [
      "SSL_CERT_FILE=${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt"
      "SSL_CERT_DIR=${pkgs.cacert}/etc/ssl/certs/"
    ];
  };
}
