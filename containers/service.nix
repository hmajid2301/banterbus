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
    paths = [package];
    pathsToLink = ["/bin"];
  };
  config = {
    ExposedPorts = {
      "8080/tcp" = {};
    };
    Env = [
      "BANTERBUS_DB_FOLDER=/"
    ];
    Cmd = ["${package}/bin/banterbus"];
  };
}
