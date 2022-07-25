let
  pkgs = import (fetchTarball {
        url = https://github.com/NixOS/nixpkgs/archive/99c46c5bb5ab0e74a94945fe9222132757667d17.tar.gz;
        sha256 = "0dhxl001a9n6hvbrdcp4bxx5whd54jyamszskknkmy65d5gbnqxd";
      }) {};
in
  pkgs.mkShell {
    nativeBuildInputs = [ pkgs.skopeo pkgs.go pkgs.cfssl ];
    KUBECONFIG = "${ builtins.toString ./testauto/kubeconfig.yaml}";
    REGISTRYMAN_REV = "f4400940fb731cd4194a66aebfb93b5823d74f95";
  }
