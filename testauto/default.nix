{ pkgs ? import (fetchTarball {
  url =
    "https://github.com/NixOS/nixpkgs/archive/99c46c5bb5ab0e74a94945fe9222132757667d17.tar.gz";
  sha256 = "0dhxl001a9n6hvbrdcp4bxx5whd54jyamszskknkmy65d5gbnqxd";
}) { }, registryman-local-path ? (toString ../.), docker-image-name ? "testauto"
, }:

let
  registryman =
    import (/. + registryman-local-path + "/default.nix") { inherit pkgs; };

  # registryman-docker-image = registryman.dockerimage;

  image-tag = builtins.hashString "sha256"
    ((builtins.toString registryman.binary) + (builtins.toString source));

  racket-base-pkgs = [
    "${racketYaml}"
    "${racketSrfi}/srfi-lite-lib"
    "${racketMzScheme}/mzscheme-lib"
    "${racketTyped}/typed-racket-lib"
    "${racketTyped}/source-syntax"
    "${racketSchemeLib}/scheme"
    "${racketStringConstants}/string-constants-lib"
    "${racketPconvert}/pconvert-lib"
    "${racketCompatLib}/compatibility-lib"
    "${racketNet}/net-lib"
    "${racketRackUnit}/rackunit-lib"
    "${racketRackUnit}/testing-util-lib"
    "${racketSandboxLib}"
    "${racketErrorTrace}/errortrace-lib"
    "${racketRash}/rash"
    "${racketRash}/linea"
    "${racketRash}/shell-pipeline"
    "${racketUdelim}"
    "${racketBasedir}/basedir"
    "${racketReadline}/readline-lib"
    "${racketPeg}/peg"
  ];

  nginx-deploy = pkgs.runCommand "testauto-nginx-deploy" {
    src = fetchTarball {
      url =
        "https://github.com/kubernetes/ingress-nginx/archive/refs/tags/controller-v1.0.4.tar.gz";
      sha256 = "09vj807my6nbihy6lf0q64v4bsq2l3pp05llz46ml6cxybvz2yca";
    };
  } ''
    mkdir -p $out
    cp -a $src/deploy/static/provider/kind/deploy.yaml $out/
  '';

  racketYaml = fetchTarball {
    url =
      "https://github.com/esilkensen/yaml/archive/b60a1e4a01979ed447799b07e7f8dd5ff17019f0.tar.gz";
    sha256 = "01r8lhz8b31fd4m5pr5ifmls1rk0rs7yy3mcga3k5wfzkvjhc6pg";
  };

  racketSrfi = fetchTarball {
    url = "https://github.com/racket/srfi/archive/refs/tags/v8.2.tar.gz";
    sha256 = "19ywzx1km5k7fzcvb2r7ymg0zh3344k7fr4cq8c817vzkz0327wp";
  };

  racketSchemeLib = fetchTarball {
    url = "https://github.com/racket/scheme-lib/archive/refs/tags/v8.2.tar.gz";
    sha256 = "0qw2yv3bmvax897n6i4zhhjpfgg5hclqjsdmfviz47gz38nsa2hr";
  };

  racketCompatLib = fetchTarball {
    url =
      "https://github.com/racket/compatibility/archive/refs/tags/v8.2.tar.gz";
    sha256 = "154k8nxnm102xx03gvrj0kybhswdn0g9gnqfp5lly2vrx2li3mc8";
  };

  racketMzScheme = fetchTarball {
    url = "https://github.com/racket/mzscheme/archive/refs/tags/v8.2.tar.gz";
    sha256 = "1cbgg8j7di9fhm5ll4429jddch4rpifgaz3h32ddkvkslibfqfpb";
  };

  racketTyped = fetchTarball {
    url =
      "https://github.com/racket/typed-racket/archive/refs/tags/v8.2.tar.gz";
    sha256 = "1zjca6s5pf125liqjms8yfhl30p4wl2mnhbkkavkfxrpd2p00rch";
  };

  racketStringConstants = fetchTarball {
    url =
      "https://github.com/racket/string-constants/archive/refs/tags/v8.2.tar.gz";
    sha256 = "1m8h61vpcc1x2ag18fwz65lglzq68llxkfxqr3aj77nixhkk96dm";
  };

  racketPconvert = fetchTarball {
    url = "https://github.com/racket/pconvert/archive/refs/tags/v8.2.tar.gz";
    sha256 = "0sk2cwz1a3s59a1cck6cdg6zmmcj8ix9fmm9d211ly25nbk87hif";
  };

  racketNet = fetchTarball {
    url = "https://github.com/racket/net/archive/refs/tags/v6.5.tar.gz";
    sha256 = "0jcj9safpnk7w8hji377sr272542jzyk7jnyzdj8q7pbd13fh026";
  };

  racketSandboxLib = fetchTarball {
    url = "https://github.com/racket/sandbox-lib/archive/refs/tags/v8.2.tar.gz";
    sha256 = "0h0amyry93h0qik6p5j4dmlg9bxlyrwj0f34bfiv8ba1933494np";
  };

  racketRackUnit = fetchTarball {
    url = "https://github.com/racket/rackunit/archive/refs/tags/v8.2.tar.gz";
    sha256 = "0yrlhkz4xn2q19ysm9psnfr5lqqdvdqh8dcdm6k6jm2sq8vzsqyh";
  };

  racketErrorTrace = fetchTarball {
    url = "https://github.com/racket/errortrace/archive/refs/tags/v8.2.tar.gz";
    sha256 = "0pgq2c0j9pyrq1hfc9mqjfwid3lcrjdvczsv57bmr60vpy5dhyai";
  };

  racketRash = fetchTarball {
    url =
      "https://github.com/willghatch/racket-rash/archive/c40c5adfedf632bc1fdbad3e0e2763b134ee3ff5.tar.gz";
    sha256 = "1jcdlidbp1nq3jh99wsghzmyamfcs5zwljarrwcyfnkmkaxvviqg";
  };

  racketUdelim = fetchTarball {
    url =
      "https://github.com/willghatch/racket-udelim/archive/58420f53c37e0bee451daa3dc5e2d72f7fc4d967.tar.gz";
    sha256 = "0h3ha4qxh8jhxg1phyqnbz51xznzgjgfxaaxxxj1wp2kdy3dn7ff";
  };

  racketBasedir = pkgs.runCommand "racketBasedir" {
    src = fetchTarball {
      url =
        "https://github.com/willghatch/racket-basedir/archive/ef95b1eeb9b4e0df491680e5caa98eeadf64dfa1.tar.gz";
      sha256 = "0xdy48mp86mi0ymz3a28vkr4yc6gid32nkjvdkhz81m5v51yxa9s";
    };
  } ''
    mkdir -p $out/basedir
    cp -a $src/* $out/basedir
  '';

  racketReadline = fetchTarball {
    url = "https://github.com/racket/readline/archive/refs/tags/v8.2.tar.gz";
    sha256 = "183pzndry2iq5z1j16yjvyin90gz70ymxbhvkr6z4l0kmmpm38x5";
  };

  racketPeg = pkgs.runCommand "racketPeg" {
    src = fetchTarball {
      url =
        "https://github.com/rain-1/racket-peg/archive/5d282d711f2b6655a4de313b603c47dddfb40d27.tar.gz";
      sha256 = "15xw3zynw3kcq151jinmsc77rpd73qgb9fi92yxr574d9y1ga5l3";
    };
  } ''
    mkdir -p $out/peg
    cp -a $src/* $out/peg
  '';

  harborHelmChart_1_6_4 = fetchTarball {
    url =
      "https://github.com/goharbor/harbor-helm/archive/refs/tags/v1.6.4.tar.gz";
    sha256 = "1424g9h5acl9qpm2jlfs5cpwrq9v93hq7siv8g80ds7yi3lwryq2";
  };

  harborHelmChart_1_7_3 = fetchTarball {
    url =
      "https://github.com/goharbor/harbor-helm/archive/refs/tags/v1.7.3.tar.gz";
    sha256 = "16qv8idkcnba7fdcj9jd8qlff95l5w535yay6g39zhmkaj0v23sv";
  };

  harborHelmChart_1_9_3 = fetchTarball {
    url =
      "https://github.com/goharbor/harbor-helm/archive/refs/tags/v1.9.3.tar.gz";
    sha256 = "1jvjn7yxxmwkwszwajyfjh9p47sb7lw9dihhhzqqjzds3sx66ain";
  };

  source = pkgs.runCommand "testauto-source" {
    src = pkgs.lib.sourceByRegex ./. [ ".*.rkt$" "harbor-values.yaml" ];
  } ''
    mkdir -p $out/testauto
    cp -a $src/* $out/testauto/
  '';

  plt-user-home = pkgs.runCommand "plt-user-home" {
    buildInputs = [ pkgs.racket-minimal ];
  } ''
    export HOME=$out
    raco pkg install --copy --no-docs --user --batch --deps force --force ${
      toString racket-base-pkgs
    }
  '';

  cert-manager-yaml = pkgs.fetchurl {
    url =
      "https://github.com/jetstack/cert-manager/releases/download/v1.6.1/cert-manager.yaml";
    sha256 = "1lix59cclkn7n43agnr36f0dmwlsca4cbs2vbahfqck04fy0jxn9";
  };

  docker = pkgs.dockerTools.buildLayeredImage {
    name = docker-image-name;
    tag = image-tag;
    contents = [
      pkgs.kind
      pkgs.docker-client
      pkgs.kubernetes-helm
      pkgs.kubectl
      pkgs.pigz
      pkgs.kustomize
    ];
    config = {
      Env = [
        "HOME=${plt-user-home}"
        "DOCKER=1"
        "REGISTRYMAN=${registryman.binary}/bin/registryman"
        "PROJECT_CRD=${registryman.project-crd}"
        "REGISTRY_CRD=${registryman.registry-crd}"
        "SCANNER_CRD=${registryman.scanner-crd}"
        "CERT_MANAGER_YAML=${cert-manager-yaml}"
        "REGISTRYMAN_DEPLOYMENT_MANIFESTS=${registryman.deployment-manifests}"
        "REGISTRYMAN_DEPLOYMENT_MANIFESTS_VERBOSE=${registryman.deployment-manifests-verbose}"
        "PLTCOLLECTS=${source}"
      ];
      Entrypoint = [ "${pkgs.racket-minimal}/bin/racket" "-l" "testauto" "--" ];
    };
  };

  trivy = import ./trivy.nix { inherit pkgs; };

  shell = pkgs.mkShell ({
    HOME = "${plt-user-home}";
    REGISTRYMAN = "${registryman.binary}/bin/registryman";
    HARBOR_VALUES_FILE = "${source}/testauto/harbor-values.yaml";
    HARBOR_HELM_1_6_4 = harborHelmChart_1_6_4;
    HARBOR_HELM_1_7_3 = harborHelmChart_1_7_3;
    HARBOR_HELM_1_9_3 = harborHelmChart_1_9_3;
    TRIVY_HELM = trivy.trivy-helm;
    TRIVY_VALUES = trivy.trivy-values;
    PROJECT_CRD = registryman.project-crd;
    REGISTRY_CRD = registryman.registry-crd;
    SCANNER_CRD = registryman.scanner-crd;
    CERT_MANAGER_YAML = cert-manager-yaml;
    NGINX_DEPLOY = "${nginx-deploy}/deploy.yaml";
    REGISTRYMAN_DOCKER_IMAGE = registryman.dockerimage;
    REGISTRYMAN_DEPLOYMENT_MANIFESTS = registryman.deployment-manifests;
    REGISTRYMAN_DEPLOYMENT_MANIFESTS_VERBOSE =
      registryman.deployment-manifests-verbose;
    nativeBuildInputs = [
      pkgs.racket-minimal
      pkgs.kind
      pkgs.docker-client
      pkgs.kubernetes-helm
      pkgs.kubectl
      pkgs.pigz
      pkgs.kustomize
    ];
    PLTCOLLECTS = ":${source}";
  });

  testauto-script = pkgs.runCommand "testauto-script" { } ''
    mkdir -p $out/bin
    cp -a ${./ta.sh} $out/bin/testauto
  '';

in {
  inherit shell docker image-tag source plt-user-home nginx-deploy
    cert-manager-yaml harborHelmChart_1_6_4 harborHelmChart_1_7_3
    harborHelmChart_1_9_3 testauto-script;
}
