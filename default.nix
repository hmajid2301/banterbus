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
  buildGoApplication ? pkgs.buildGoApplication,
}:
buildGoApplication {
  pname = "banterbus";
  version = "0.0.1";
  go = pkgs.go_1_22;
  pwd = ./.;
  src = ./.;
  modules = ./gomod2nix.toml;
  # INFO: Turn off tests as we run them in CI
  checkPhase = false;
}
