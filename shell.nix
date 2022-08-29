let
  pkgs = import (fetchTarball {
    url =
      "https://github.com/NixOS/nixpkgs/archive/99c46c5bb5ab0e74a94945fe9222132757667d17.tar.gz";
    sha256 = "0dhxl001a9n6hvbrdcp4bxx5whd54jyamszskknkmy65d5gbnqxd";
  }) { };
  registryman = import ./default.nix { inherit pkgs; };
  testauto = import ./testauto/default.nix { inherit pkgs; };
in pkgs.mkShell {
  packages = with pkgs;
    [
      skopeo
      go
      cfssl
      bash
      k9s
      kind
      kubectl
      kubernetes-helm
      kustomize
      pigz
      racket-minimal
      testauto.testauto-script
    ] ++ (with registryman; [ controller-tools code-generator binary ]);
  shellHook = ''
    echo "Setting up the registryman development environment"
    rm ./manifests
    ln -s ${registryman.deployment-manifests} ./manifests
    rm ./vendor
    ln -s ${registryman.vendor} ./vendor
  '';
  KUBECONFIG = "${builtins.toString ./testauto/kubeconfig.yaml}";
  PLTCOLLECTS = ":${testauto.source}";
  PLTUSERHOME = builtins.toString testauto.plt-user-home;
  NGINX_DEPLOY = "${testauto.nginx-deploy}/deploy.yaml";
  CERT_MANAGER_YAML = builtins.toString testauto.cert-manager-yaml;
  REGISTRYMAN_DOCKER_IMAGE = builtins.toString registryman.dockerimage;
  PROJECT_CRD = builtins.toString registryman.project-crd;
  REGISTRY_CRD = builtins.toString registryman.registry-crd;
  SCANNER_CRD = builtins.toString registryman.scanner-crd;
  REGISTRYMAN_DEPLOYMENT_MANIFESTS =
    builtins.toString registryman.deployment-manifests;
  REGISTRYMAN_DEPLOYMENT_MANIFESTS_VERBOSE =
    builtins.toString registryman.deployment-manifests-verbose;
  HARBOR_VALUES_FILE = "${testauto.source}/testauto/harbor-values.yaml";
  HARBOR_HELM_1_6_4 = builtins.toString testauto.harborHelmChart_1_6_4;
  HARBOR_HELM_1_7_3 = builtins.toString testauto.harborHelmChart_1_7_3;
  HARBOR_HELM_1_9_3 = builtins.toString testauto.harborHelmChart_1_9_3;
  REGISTRYMAN = "${registryman.binary}/bin/registryman";
}
