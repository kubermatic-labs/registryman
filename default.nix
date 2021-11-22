{
  pkgs ?
      import (fetchTarball {
        url = https://github.com/NixOS/nixpkgs/archive/34ad3ffe08adfca17fcb4e4a47bb5f3b113687be.tar.gz;
        sha256 = "02li241rz5668nfyp88zfjilxf0mr9yansa93fbl38hjwkhf3ix6";
      }) {},
  registryman-git-rev ? "",
  registryman-git-ref ? "",
  registryman-git-url ? "git@github.com:origoss/registryman.git",
  local-vendor-sha256 ? "0ydxasq4a6ln1kkc25g9jp78dk2mfg9j06qrmcaaiyq8vzb0fmk6",
  git-vendor-sha256 ? "0gcxhzi24ali7kn9433igmzkw36yf7svvgnvv2jc7xsfhg165p63",
  registryman-from ? "local",
}:

assert registryman-from == "local" || registryman-from == "git";
assert registryman-from == "git" -> registryman-git-rev != "";
assert registryman-from == "git" -> registryman-git-url != "";

let
  registryman-local-source = pkgs.runCommand "registryman-local-source" {
    src = pkgs.nix-gitignore.gitignoreSource [
      "*.nix"
      "rm-dev.sh"
      "rm-git.sh"
      "docker-build.sh"
      "update-code.sh"
      "result"
      "generated-code"
    ] ./.;
    nativeBuildInputs = [ pkgs.cpio ];
  } ''
    mkdir -p $out
    cp -a $src/* $out/
  '';

  registryman-vendor = registryman-src: sha256:
    pkgs.runCommand "registryman-local-vendor" {
      nativeBuildInputs = [ pkgs.go pkgs.cacert pkgs.git ];
      impureEnvVars = pkgs.lib.fetchers.proxyImpureEnvVars ++ [
        "GIT_PROXY_COMMAND" "SOCKS_SERVER"
      ];
      outputHashMode = "recursive";
      outputHashAlgo = "sha256";
      outputHash = sha256;
    } ''
       mkdir -p $out
       mkdir -p tmp
       cp -a ${registryman-src}/* $TMPDIR/tmp
       chmod a+w $TMPDIR/tmp/go.mod
       cd tmp
       go mod vendor
       mv vendor/* $out
    '';

  registryman-local-vendor = registryman-vendor registryman-local-source local-vendor-sha256;

  registryman-local-generated = pkgs.runCommand "registryman-local-generated" {
  } ''
      mkdir $out
      cp -a ${registryman-local-source}/* $TMPDIR
      chmod a+rw -R $TMPDIR
      cp -a ${generated-code}/* $TMPDIR/pkg/apis/registryman/v1alpha1/
      cp -a $TMPDIR/* $out
    '';

#   registryman-local-generated = pkgs.runCommand "registryman-local-generated" {
#     nativeBuildInputs = [ pkgs.go ];
#   } ''
#     mkdir -p $out
#     # cp ${registryman-local-source}/* $out
#     # cp ${generated-code}/* $out/pkg/apis/registryman/v1alpha1/
#     # mkdir -p $TMPDIR/go/src/github.com/kubermatic-labs/registryman
#     # cp -a ${registryman-local-source}/* $TMPDIR/go/src/github.com/kubermatic-labs/registryman
#     # ln -s ${registryman-local-vendor} $TMPDIR/go/src/github.com/kubermatic-labs/registryman/vendor
#     # cd $TMPDIR/go/src/github.com/kubermatic-labs/registryman
#     # chmod a+rw -R .
#     # patchShebangs --build $TMPDIR/go/src/github.com/kubermatic-labs/registryman/hack/update-codegen.sh
#     # export GOPATH=$TMPDIR/go
#     # hack/update-codegen.sh
#     # rm $TMPDIR/go/src/github.com/kubermatic-labs/registryman/vendor
#     # mv $TMPDIR/go/src/github.com/kubermatic-labs/registryman/* $out
# '';

  registryman = registryman-vendor:
    pkgs.runCommand "registryman-local" {
      buildInputs = [ pkgs.go pkgs.removeReferencesTo ];
      disallowedReferences = [ pkgs.go ];
    } ''
       mkdir -p $out/bin
       mkdir -p $TMPDIR/go/src/github.com/kubermatic-labs/registryman
       cp -a ${registryman-generated}/* $TMPDIR/go/src/github.com/kubermatic-labs/registryman
       ln -s ${registryman-vendor} $TMPDIR/go/src/github.com/kubermatic-labs/registryman/vendor
       cd $TMPDIR/go/src/github.com/kubermatic-labs/registryman
       export GOPATH=$TMPDIR/go
       export CGO_ENABLED=0
       go build -tags "exclude_graphdriver_devicemapper exclude_graphdriver_btrfs containers_image_openpgp" -ldflags="-s -w"
       remove-references-to -t ${pkgs.go} $TMPDIR/go/src/github.com/kubermatic-labs/registryman/registryman
       mv $TMPDIR/go/src/github.com/kubermatic-labs/registryman/registryman $out/bin
   '';

  registryman-git-source = fetchGit {
      url = registryman-git-url;
      ref = registryman-git-ref;
      rev = registryman-git-rev;
    };

  registryman-git-vendor = registryman-vendor registryman-git-source git-vendor-sha256;

  registryman-generated = if registryman-from == "local" then registryman-local-generated
    else registryman-git-source;

  registryman-built = if registryman-from == "local" then
    registryman registryman-local-vendor else
      registryman registryman-git-vendor;

  dockerimage = registryman-pkg: pkgs.dockerTools.buildLayeredImage {
    name = "registryman";
    config = {
      Entrypoint = [ "${registryman-pkg}/bin/registryman" ];
    };
  };

  controller-tools = pkgs.buildGoModule rec {
    pname = "controller-tools";
    version = "0.7.0";

    doCheck = false;
    src = pkgs.fetchFromGitHub {
      owner = "kubernetes-sigs";
      repo = "controller-tools";
      rev = "v${version}";
      sha256 = "089iz2g4xj7b5cgmjd9xp1l30k5lbnibiiqfcr659rjprbv1yv1f";
    };

    vendorSha256 = "1p8hx3a62l4drjba8wg2frwvm369lls2d2yab74knb109d0g2v51";
  };

  code-generator = pkgs.buildGoModule rec {
    pname = "code-generator";
    version = "0.22.4";

    # doCheck = false;
    src = pkgs.fetchFromGitHub {
      owner = "kubernetes";
      repo = "code-generator";
      rev = "v${version}";
      sha256 = "09z3wrpjxiyqbx3djryrwkq048npqnj2hrmybbmywgdm9z9v70i4";
    };

    vendorSha256 = "1gsva0z8dc0yild046b761kqhhh1g0dqs6qkcqlnl2mvgzwdahx6";
  };

  collected-go-sources = pkgs.runCommand "collected-go-sources" {
  } ''
     mkdir -p $out/
     mkdir -p $TMPDIR/go
     cp -a ${pkgs.go}/share/go/* $TMPDIR/go
     chmod a+rw -R $TMPDIR/go
     cp -a ${registryman-local-vendor}/* $TMPDIR/go/src
     chmod a+rw -R $TMPDIR/go
     mkdir -p $TMPDIR/go/src/github.com/kubermatic-labs/registryman
     cp -a ${registryman-local-source}/* $TMPDIR/go/src/github.com/kubermatic-labs/registryman
     chmod a+rw -R $TMPDIR/go/src/github.com/kubermatic-labs/registryman
     mv $TMPDIR/go/* $out
  '';

  generated-code = pkgs.runCommand "registryman-generated" {
    nativeBuildInputs = [controller-tools code-generator pkgs.go ];
    GO111MODULE = "on";
    GOROOT = collected-go-sources;
  } ''
    mkdir -p $out/
    mkdir -p $TMPDIR/go/src/github.com/kubermatic-labs/registryman
    cp -a ${registryman-local-source}/* $TMPDIR/go/src/github.com/kubermatic-labs/registryman
    ln -s ${registryman-local-vendor} $TMPDIR/go/src/github.com/kubermatic-labs/registryman/vendor
    cd $TMPDIR/go/src/github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1
    deepcopy-gen -i github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1 \
                 -O zz_generated.deepcopy                                                \
                 -h ${collected-go-sources}/src/github.com/kubermatic-labs/registryman/hack/boilerplate.go.txt \
                 -o $TMPDIR/generated
    client-gen -i github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1 \
               -h ${collected-go-sources}/src/github.com/kubermatic-labs/registryman/hack/boilerplate.go.txt \
               --input "registryman/v1alpha1" \
               --input-base "github.com/kubermatic-labs/registryman/pkg/apis" \
               -p github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1/clientset \
               --clientset-name versioned \
               -o $TMPDIR/generated
    lister-gen -i github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1 \
               -h ${collected-go-sources}/src/github.com/kubermatic-labs/registryman/hack/boilerplate.go.txt \
               -p github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1/listers \
               -o $TMPDIR/generated
    informer-gen -i github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1 \
                 -h ${collected-go-sources}/src/github.com/kubermatic-labs/registryman/hack/boilerplate.go.txt \
                 -p github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1/informers \
                 --versioned-clientset-package github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1/clientset/versioned \
                 --listers-package github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1/listers \
                 -o $TMPDIR/generated
    register-gen -i github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1 \
                 -h ${collected-go-sources}/src/github.com/kubermatic-labs/registryman/hack/boilerplate.go.txt \
                 -p github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1/informers \
                 -o $TMPDIR/generated
    openapi-gen -i github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1 \
                -h ${collected-go-sources}/src/github.com/kubermatic-labs/registryman/hack/boilerplate.go.txt \
                -p github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1 \
                -o $TMPDIR/generated
    controller-gen crd output:dir=$TMPDIR/generated/github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1
    mv $TMPDIR/generated/github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1/* $out
  '';

in {
  # inherit registryman-local registryman-git;
  inherit registryman-built generated-code;

  shell = pkgs.mkShell {
    nativeBuildInputs = [ registryman-built ];
  };

  dev = pkgs.mkShell {
    nativeBuildInputs = [ controller-tools code-generator pkgs.go ];
    REGISTRYMAN_SRC = registryman-local-source;
    REGISTRYMAN_VENDOR = registryman-local-vendor;
    GO111MODULE = "on";
    GOROOT = collected-go-sources;
    REGISTRYMAN_GENERATED = generated-code;
  };
  docker = dockerimage registryman-built;

  project-crd = "${registryman-generated}/pkg/apis/registryman/v1alpha1/registryman.kubermatic.com_projects.yaml";
  registry-crd = "${registryman-generated}/pkg/apis/registryman/v1alpha1/registryman.kubermatic.com_registries.yaml";
  scanner-crd = "${registryman-generated}/pkg/apis/registryman/v1alpha1/registryman.kubermatic.com_scanners.yaml";
}
