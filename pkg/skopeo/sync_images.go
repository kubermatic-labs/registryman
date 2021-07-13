/*
   Copyright 2021 The Kubermatic Kubernetes Platform contributors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package skopeo

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/directory"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/docker/reference"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports"
	"github.com/containers/image/v5/types"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type repoDescriptor struct {
	DirBasePath string                 // base path when source is 'dir'
	ImageRefs   []types.ImageReference // List of tagged image found for the repository
	Context     *types.SystemContext   // SystemContext for the sync command
}

// tlsVerifyConfig is an implementation of the Unmarshaler interface, used to
// customize the unmarshaling behaviour of the tls-verify YAML key.
type tlsVerifyConfig struct {
	skip types.OptionalBool // skip TLS verification check (false by default)
}

// registrySyncConfig contains information about a single registry, read from
// the source YAML file
type registrySyncConfig struct {
	Images           map[string][]string    // Images map images name to slices with the images' references (tags, digests)
	ImagesByTagRegex map[string]string      `yaml:"images-by-tag-regex"` // Images map images name to regular expression with the images' tags
	Credentials      types.DockerAuthConfig // Username and password used to authenticate with the registry
	TLSVerify        tlsVerifyConfig        `yaml:"tls-verify"` // TLS verification mode (enabled by default)
	CertDir          string                 `yaml:"cert-dir"`   // Path to the TLS certificates of the registry
}

// sourceConfig contains all registries information read from the source YAML file
type sourceConfig map[string]registrySyncConfig

func syncImages(data *transferData) error {

	// validate source and destination options
	contains := func(val string, list []string) (_ bool) {
		for _, l := range list {
			if l == val {
				return true
			}
		}
		return
	}

	ctx := context.Background()
	policy := &signature.Policy{
		Default: []signature.PolicyRequirement{
			signature.NewPRInsecureAcceptAnything(),
		},
	}
	policyContext, err := signature.NewPolicyContext(policy)
	if err != nil {
		return err
	}

	if len(data.sourceTransport) == 0 {
		return errors.New("A source transport must be specified")
	}
	if !contains(data.sourceTransport, []string{docker.Transport.Name(), directory.Transport.Name(), "yaml"}) {
		return errors.Errorf("%q is not a valid source transport", data.sourceTransport)
	}

	if len(data.destinationTransport) == 0 {
		return errors.New("A destination transport must be specified")
	}
	if !contains(data.destinationTransport, []string{docker.Transport.Name(), directory.Transport.Name()}) {
		return errors.Errorf("%q is not a valid destination transport", data.destinationTransport)
	}

	if data.sourceTransport == data.destinationTransport && data.sourceTransport == directory.Transport.Name() {
		return errors.New("sync from 'dir' to 'dir' not implemented, consider using rsync instead")
	}

	srcRepoList, err := imagesToCopy(data.sourcePath, data.sourceTransport, data.sourceCtx)
	if err != nil {
		return err
	}

	imageListSelection := copy.CopyAllImages
	imagesNumber := 0
	options := copy.Options{
		ReportWriter:       os.Stdout,
		DestinationCtx:     data.destinationCtx,
		ImageListSelection: imageListSelection,
	}

	for _, srcRepo := range srcRepoList {
		options.SourceCtx = srcRepo.Context
		for counter, ref := range srcRepo.ImageRefs {
			var destSuffix string
			switch ref.Transport() {
			case docker.Transport:
				// docker -> dir or docker -> docker
				destSuffix = ref.DockerReference().String()
			case directory.Transport:
				// dir -> docker (we don't allow `dir` -> `dir` sync operations)
				destSuffix = strings.TrimPrefix(ref.StringWithinTransport(), srcRepo.DirBasePath)
				if destSuffix == "" {
					// if source is a full path to an image, have destPath scoped to repo:tag
					destSuffix = path.Base(srcRepo.DirBasePath)
				}
			}

			if !data.scoped {
				destSuffix = path.Base(destSuffix)
			}

			destRef, err := destinationReference(path.Join(data.destinationPath, destSuffix), data.destinationTransport)
			if err != nil {
				return err
			}

			logrus.WithFields(logrus.Fields{
				"from": transports.ImageName(ref),
				"to":   transports.ImageName(destRef),
			}).Infof("Copying image ref %d/%d", counter+1, len(srcRepo.ImageRefs))

			_, err = copy.Image(ctx, policyContext, destRef, ref, &options)

			if err != nil {
				return errors.Wrapf(err, "Error copying ref %q", transports.ImageName(ref))
			}
			imagesNumber++
		}
	}
	logrus.Infof("Synced %d images from %d sources", imagesNumber, len(srcRepoList))
	return nil
}

// imagesToCopy retrieves all the images to copy from a specified sync source
// and transport.
// It returns a slice of repository descriptors, where each descriptor is a
// list of tagged image references to be used as sync source, and any error
// encountered.
func imagesToCopy(source string, transport string, sourceCtx *types.SystemContext) ([]repoDescriptor, error) {
	var descriptors []repoDescriptor

	switch transport {
	case docker.Transport.Name():
		desc := repoDescriptor{
			Context: sourceCtx,
		}
		named, err := reference.ParseNormalizedNamed(source) // May be a repository or an image.
		if err != nil {
			return nil, errors.Wrapf(err, "Cannot obtain a valid image reference for transport %q and reference %q", docker.Transport.Name(), source)
		}
		imageTagged := !reference.IsNameOnly(named)
		logrus.WithFields(logrus.Fields{
			"imagename": source,
			"tagged":    imageTagged,
		}).Info("Tag presence check")
		if imageTagged {
			srcRef, err := docker.NewReference(named)
			if err != nil {
				return nil, errors.Wrapf(err, "Cannot obtain a valid image reference for transport %q and reference %q", docker.Transport.Name(), named.String())
			}
			desc.ImageRefs = []types.ImageReference{srcRef}
		} else {
			desc.ImageRefs, err = imagesToCopyFromRepo(sourceCtx, named)
			if err != nil {
				return descriptors, err
			}
			if len(desc.ImageRefs) == 0 {
				return descriptors, errors.Errorf("No images to sync found in %q", source)
			}
		}
		descriptors = append(descriptors, desc)

	case directory.Transport.Name():
		desc := repoDescriptor{
			Context: sourceCtx,
		}

		if _, err := os.Stat(source); err != nil {
			return descriptors, errors.Wrap(err, "Invalid source directory specified")
		}
		desc.DirBasePath = source
		var err error
		desc.ImageRefs, err = imagesToCopyFromDir(source)
		if err != nil {
			return descriptors, err
		}
		if len(desc.ImageRefs) == 0 {
			return descriptors, errors.Errorf("No images to sync found in %q", source)
		}
		descriptors = append(descriptors, desc)

	case "yaml":
		cfg, err := newSourceConfig(source)
		if err != nil {
			return descriptors, err
		}
		for registryName, registryConfig := range cfg {
			if len(registryConfig.Images) == 0 && len(registryConfig.ImagesByTagRegex) == 0 {
				logrus.WithFields(logrus.Fields{
					"registry": registryName,
				}).Warn("No images specified for registry")
				continue
			}

			descs, err := imagesToCopyFromRegistry(registryName, registryConfig, *sourceCtx)
			if err != nil {
				return descriptors, errors.Wrapf(err, "Failed to retrieve list of images from registry %q", registryName)
			}
			descriptors = append(descriptors, descs...)
		}
	}

	return descriptors, nil
}

// newSourceConfig unmarshals the provided YAML file path to the sourceConfig type.
// It returns a new unmarshaled sourceConfig object and any error encountered.
func newSourceConfig(yamlFile string) (sourceConfig, error) {
	var cfg sourceConfig
	source, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(source, &cfg)
	if err != nil {
		return cfg, errors.Wrapf(err, "Failed to unmarshal %q", yamlFile)
	}
	return cfg, nil
}

// imagesToCopyFromRepo builds a list of image references from the tags
// found in a source repository.
// It returns an image reference slice with as many elements as the tags found
// and any error encountered.
func imagesToCopyFromRepo(sys *types.SystemContext, repoRef reference.Named) ([]types.ImageReference, error) {
	tags, err := getImageTags(context.Background(), sys, repoRef)
	if err != nil {
		return nil, err
	}

	var sourceReferences []types.ImageReference
	for _, tag := range tags {
		taggedRef, err := reference.WithTag(repoRef, tag)
		if err != nil {
			return nil, errors.Wrapf(err, "Error creating a reference for repository %s and tag %q", repoRef.Name(), tag)
		}
		ref, err := docker.NewReference(taggedRef)
		if err != nil {
			return nil, errors.Wrapf(err, "Cannot obtain a valid image reference for transport %q and reference %s", docker.Transport.Name(), taggedRef.String())
		}
		sourceReferences = append(sourceReferences, ref)
	}
	return sourceReferences, nil
}

// imagesToCopyFromDir builds a list of image references from the images found
// in the source directory.
// It returns an image reference slice with as many elements as the images found
// and any error encountered.
func imagesToCopyFromDir(dirPath string) ([]types.ImageReference, error) {
	var sourceReferences []types.ImageReference
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == "manifest.json" {
			dirname := filepath.Dir(path)
			ref, err := directory.Transport.ParseReference(dirname)
			if err != nil {
				return errors.Wrapf(err, "Cannot obtain a valid image reference for transport %q and reference %q", directory.Transport.Name(), dirname)
			}
			sourceReferences = append(sourceReferences, ref)
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		return sourceReferences,
			errors.Wrapf(err, "Error walking the path %q", dirPath)
	}

	return sourceReferences, nil
}

// imagesToCopyFromRegistry builds a list of repository descriptors from the images
// in a registry configuration.
// It returns a repository descriptors slice with as many elements as the images
// found and any error encountered. Each element of the slice is a list of
// image references, to be used as sync source.
func imagesToCopyFromRegistry(registryName string, cfg registrySyncConfig, sourceCtx types.SystemContext) ([]repoDescriptor, error) {
	serverCtx := &sourceCtx
	// override ctx with per-registryName options
	serverCtx.DockerCertPath = cfg.CertDir
	serverCtx.DockerDaemonCertPath = cfg.CertDir
	serverCtx.DockerDaemonInsecureSkipTLSVerify = (cfg.TLSVerify.skip == types.OptionalBoolTrue)
	serverCtx.DockerInsecureSkipTLSVerify = cfg.TLSVerify.skip
	if cfg.Credentials != (types.DockerAuthConfig{}) {
		serverCtx.DockerAuthConfig = &cfg.Credentials
	}
	var repoDescList []repoDescriptor
	for imageName, refs := range cfg.Images {
		repoLogger := logrus.WithFields(logrus.Fields{
			"repo":     imageName,
			"registry": registryName,
		})
		repoRef, err := parseRepositoryReference(fmt.Sprintf("%s/%s", registryName, imageName))
		if err != nil {
			repoLogger.Error("Error parsing repository name, skipping")
			logrus.Error(err)
			continue
		}

		repoLogger.Info("Processing repo")

		var sourceReferences []types.ImageReference
		if len(refs) != 0 {
			for _, ref := range refs {
				tagLogger := logrus.WithFields(logrus.Fields{"ref": ref})
				var named reference.Named
				// first try as digest
				if d, err := digest.Parse(ref); err == nil {
					named, err = reference.WithDigest(repoRef, d)
					if err != nil {
						tagLogger.Error("Error processing ref, skipping")
						logrus.Error(err)
						continue
					}
				} else {
					tagLogger.Debugf("Ref was not a digest, trying as a tag: %s", err)
					named, err = reference.WithTag(repoRef, ref)
					if err != nil {
						tagLogger.Error("Error parsing ref, skipping")
						logrus.Error(err)
						continue
					}
				}

				imageRef, err := docker.NewReference(named)
				if err != nil {
					tagLogger.Error("Error processing ref, skipping")
					logrus.Errorf("Error getting image reference: %s", err)
					continue
				}
				sourceReferences = append(sourceReferences, imageRef)
			}
		} else { // len(refs) == 0
			repoLogger.Info("Querying registry for image tags")
			sourceReferences, err = imagesToCopyFromRepo(serverCtx, repoRef)
			if err != nil {
				repoLogger.Error("Error processing repo, skipping")
				logrus.Error(err)
				continue
			}
		}

		if len(sourceReferences) == 0 {
			repoLogger.Warnf("No refs to sync found")
			continue
		}
		repoDescList = append(repoDescList, repoDescriptor{
			ImageRefs: sourceReferences,
			Context:   serverCtx})
	}

	for imageName, tagRegex := range cfg.ImagesByTagRegex {
		repoLogger := logrus.WithFields(logrus.Fields{
			"repo":     imageName,
			"registry": registryName,
		})
		repoRef, err := parseRepositoryReference(fmt.Sprintf("%s/%s", registryName, imageName))
		if err != nil {
			repoLogger.Error("Error parsing repository name, skipping")
			logrus.Error(err)
			continue
		}

		repoLogger.Info("Processing repo")

		var sourceReferences []types.ImageReference

		tagReg, err := regexp.Compile(tagRegex)
		if err != nil {
			repoLogger.WithFields(logrus.Fields{
				"regex": tagRegex,
			}).Error("Error parsing regex, skipping")
			logrus.Error(err)
			continue
		}

		repoLogger.Info("Querying registry for image tags")
		allSourceReferences, err := imagesToCopyFromRepo(serverCtx, repoRef)
		if err != nil {
			repoLogger.Error("Error processing repo, skipping")
			logrus.Error(err)
			continue
		}

		repoLogger.Infof("Start filtering using the regular expression: %v", tagRegex)
		for _, sReference := range allSourceReferences {
			tagged, isTagged := sReference.DockerReference().(reference.Tagged)
			if !isTagged {
				repoLogger.Errorf("Internal error, reference %s does not have a tag, skipping", sReference.DockerReference())
				continue
			}
			if tagReg.MatchString(tagged.Tag()) {
				sourceReferences = append(sourceReferences, sReference)
			}
		}

		if len(sourceReferences) == 0 {
			repoLogger.Warnf("No refs to sync found")
			continue
		}
		repoDescList = append(repoDescList, repoDescriptor{
			ImageRefs: sourceReferences,
			Context:   serverCtx})
	}

	return repoDescList, nil
}

// getImageTags lists all tags in a repository.
// It returns a string slice of tags and any error encountered.
func getImageTags(ctx context.Context, sysCtx *types.SystemContext, repoRef reference.Named) ([]string, error) {
	name := repoRef.Name()
	logrus.WithFields(logrus.Fields{
		"image": name,
	}).Info("Getting tags")
	// Ugly: NewReference rejects IsNameOnly references, and GetRepositoryTags ignores the tag/digest.
	// So, we use TagNameOnly here only to shut up NewReference
	dockerRef, err := docker.NewReference(reference.TagNameOnly(repoRef))
	if err != nil {
		return nil, err // Should never happen for a reference with tag and no digest
	}
	tags, err := docker.GetRepositoryTags(ctx, sysCtx, dockerRef)

	switch err := err.(type) {
	case nil:
		break
	case docker.ErrUnauthorizedForCredentials:
		// Some registries may decide to block the "list all tags" endpoint.
		// Gracefully allow the sync to continue in this case.
		logrus.Warnf("Registry disallows tag list retrieval: %s", err)
	default:
		return tags, errors.Wrapf(err, "Error determining repository tags for image %s", name)
	}

	return tags, nil
}

// parseRepositoryReference parses input into a reference.Named, and verifies that it names a repository, not an image.
func parseRepositoryReference(input string) (reference.Named, error) {
	ref, err := reference.ParseNormalizedNamed(input)
	if err != nil {
		return nil, err
	}
	if !reference.IsNameOnly(ref) {
		return nil, errors.Errorf("input names a reference, not a repository")
	}
	return ref, nil
}

// destinationReference creates an image reference using the provided transport.
// It returns a image reference to be used as destination of an image copy and
// any error encountered.
func destinationReference(destination string, transport string) (types.ImageReference, error) {
	var imageTransport types.ImageTransport

	switch transport {
	case docker.Transport.Name():
		destination = fmt.Sprintf("//%s", destination)
		imageTransport = docker.Transport
	case directory.Transport.Name():
		_, err := os.Stat(destination)
		if err == nil {
			return nil, errors.Errorf("Refusing to overwrite destination directory %q", destination)
		}
		if !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "Destination directory could not be used")
		}
		// the directory holding the image must be created here
		if err = os.MkdirAll(destination, 0755); err != nil {
			return nil, errors.Wrapf(err, "Error creating directory for image %s", destination)
		}
		imageTransport = directory.Transport
	default:
		return nil, errors.Errorf("%q is not a valid destination transport", transport)
	}
	logrus.Debugf("Destination for transport %q: %s", transport, destination)

	destRef, err := imageTransport.ParseReference(destination)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot obtain a valid image reference for transport %q and reference %q", imageTransport.Name(), destination)
	}

	return destRef, nil
}
