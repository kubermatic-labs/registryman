{
  pkgs ?
      import (fetchTarball {
        url = https://github.com/NixOS/nixpkgs/archive/34ad3ffe08adfca17fcb4e4a47bb5f3b113687be.tar.gz;
        sha256 = "02li241rz5668nfyp88zfjilxf0mr9yansa93fbl38hjwkhf3ix6";
      }) {},
  registryman-git-rev ? "",
  registryman-git-ref ? "master",
  local-vendor-sha256 ? "0gcxhzi24ali7kn9433igmzkw36yf7svvgnvv2jc7xsfhg165p63",
  git-vendor-sha256 ? "0gcxhzi24ali7kn9433igmzkw36yf7svvgnvv2jc7xsfhg165p63",
  registryman-from ? "local",
}:

assert registryman-from == "local" || registryman-from == "git";
assert registryman-from == "git" -> registryman-git-rev != "";

let
  registryman-local-source = pkgs.runCommand "registryman-local-source" {
    src = pkgs.nix-gitignore.gitignoreSource [
      "*.nix"
      "rm-dev.sh"
      "rm-git.sh"
      "docker-build.sh"
      "pkg/apis/registryman/v1alpha1/openapi_generated.go"
      "pkg/apis/registryman/v1alpha1/zz_generated.deepcopy.go"
      "pkg/apis/registryman/v1alpha1/zz_generated.register.go"
      "pkg/apis/registryman/v1alpha1/clientset"
      "pkg/apis/registryman/v1alpha1/informers"
      "pkg/apis/registryman/v1alpha1/listers"
    ] ./.;
    nativeBuildInputs = [ pkgs.cpio ];
  } ''
    mkdir -p $out
    cp -a $src/* $out/
  '';

  registryman-vendor = registryman-src: sha256:
    pkgs.runCommand "registryman-local-vendor" {
      nativeBuildInputs = [ pkgs.go pkgs.cacert ];
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
    nativeBuildInputs = [ pkgs.go ];
  } ''
    mkdir -p $out
    mkdir -p $TMPDIR/go/src/github.com/kubermatic-labs/registryman
    cp -a ${registryman-local-source}/* $TMPDIR/go/src/github.com/kubermatic-labs/registryman
    ln -s ${registryman-local-vendor} $TMPDIR/go/src/github.com/kubermatic-labs/registryman/vendor
    cd $TMPDIR/go/src/github.com/kubermatic-labs/registryman
    chmod a+rw -R .
    patchShebangs --build $TMPDIR/go/src/github.com/kubermatic-labs/registryman/hack/update-codegen.sh
    export GOPATH=$TMPDIR/go
    hack/update-codegen.sh
    rm $TMPDIR/go/src/github.com/kubermatic-labs/registryman/vendor
    mv $TMPDIR/go/src/github.com/kubermatic-labs/registryman/* $out
'';

  registryman = registryman-source: registryman-vendor:
    pkgs.runCommand "registryman-local" {
      buildInputs = [ pkgs.go pkgs.removeReferencesTo ];
      disallowedReferences = [ pkgs.go ];
    } ''
       mkdir -p $out/bin
       mkdir -p $TMPDIR/go/src/github.com/kubermatic-labs/registryman
       cp -a ${registryman-source}/* $TMPDIR/go/src/github.com/kubermatic-labs/registryman
       ln -s ${registryman-vendor} $TMPDIR/go/src/github.com/kubermatic-labs/registryman/vendor
       cd $TMPDIR/go/src/github.com/kubermatic-labs/registryman
       export GOPATH=$TMPDIR/go
       export CGO_ENABLED=0
       go build -tags "exclude_graphdriver_devicemapper exclude_graphdriver_btrfs containers_image_openpgp" -ldflags="-s -w"
       remove-references-to -t ${pkgs.go} $TMPDIR/go/src/github.com/kubermatic-labs/registryman/registryman
       mv $TMPDIR/go/src/github.com/kubermatic-labs/registryman/registryman $out/bin
   '';

  registryman-git-source = fetchGit {
      url = "git@github.com:kubermatic-labs/registryman.git";
      ref = registryman-git-ref;
      rev = registryman-git-rev;
    };

  registryman-git-vendor = registryman-vendor registryman-git-source git-vendor-sha256;

  registryman-built = if registryman-from == "local" then
    registryman registryman-local-generated registryman-local-vendor else
      registryman registryman-git-source registryman-git-vendor;

  dockerimage = registryman-pkg: pkgs.dockerTools.buildLayeredImage {
    name = "registryman";
    config = {
      Entrypoint = [ "${registryman-pkg}/bin/registryman" ];
    };
  };

in {
  # inherit registryman-local registryman-git;
  inherit registryman-built;

  shell = pkgs.mkShell {
    nativeBuildInputs = [ registryman-built ];
  };

  docker = dockerimage registryman-built;

}
