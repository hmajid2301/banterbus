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
    ];
    pathsToLink = ["/bin"];
  };
  Cmd = ["${package}/bin/banterbus"];
}
