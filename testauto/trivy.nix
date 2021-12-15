{
  pkgs
}:
let
  trivy-src = fetchTarball{
    url = https://github.com/aquasecurity/trivy/archive/refs/tags/v0.21.2.tar.gz;
    sha256 = "1bby8bairl63xxd59pwlhqbbgjxdbmwaryq52mbxyph0mb0f7ilk";
  };

  trivy-helm = pkgs.runCommand "trivy-helm" {
  } ''
    ln -s ${trivy-src}/helm/trivy $out
  '';

  trivy-values = ./trivy-values.yaml;
in
{
  inherit trivy-helm trivy-values;
}
