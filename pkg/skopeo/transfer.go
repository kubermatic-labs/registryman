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
	"fmt"

	"github.com/containers/image/v5/directory"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/types"
	"github.com/go-logr/logr"
)

type transfer struct {
	dockerCtx *types.SystemContext
	dirCtx    *types.SystemContext
}
type transferData struct {
	sourcePath           string
	destinationPath      string
	sourceCtx            *types.SystemContext
	destinationCtx       *types.SystemContext
	sourceTransport      string
	destinationTransport string
	scoped               bool
}

// New creates a new transfer struct.
func New(username, password string) *transfer {
	return &transfer{
		dockerCtx: &types.SystemContext{
			DockerAuthConfig: &types.DockerAuthConfig{
				Username: username,
				Password: password,
			},
		},
		dirCtx: &types.SystemContext{},
	}
}

// Export exports Docker repositories from a source repository to a destination path.
func (t *transfer) Export(source, destination string, logger logr.Logger) error {
	logger.Info("exporting images started")

	err := syncImages(&transferData{
		sourcePath:           source,
		destinationPath:      destination,
		sourceCtx:            t.dockerCtx,
		destinationCtx:       t.dirCtx,
		sourceTransport:      docker.Transport.Name(),
		destinationTransport: directory.Transport.Name(),
		scoped:               true,
	})

	if err != nil {
		return fmt.Errorf("syncing images failed: %w", err)
	}

	return nil
}

// Import imports Docker repositories from a source path to a destination repository.
func (t *transfer) Import(source, destination string, logger logr.Logger) error {
	logger.Info("importing images started")

	err := syncImages(&transferData{
		sourcePath:           source,
		destinationPath:      destination,
		sourceCtx:            t.dirCtx,
		destinationCtx:       t.dockerCtx,
		sourceTransport:      directory.Transport.Name(),
		destinationTransport: docker.Transport.Name(),
		scoped:               false,
	})

	if err != nil {
		return fmt.Errorf("syncing images failed: %w", err)
	}

	return nil
}

// Sync synchronizes Docker repositories from a source repository to a destination repository.
func (t *transfer) Sync(sourceRepo, destinationRepo string, logger logr.Logger) error {
	logger.Info("syncing images started")
	err := syncImages(&transferData{
		sourcePath:           sourceRepo,
		destinationPath:      destinationRepo,
		sourceCtx:            t.dockerCtx,
		destinationCtx:       t.dockerCtx,
		sourceTransport:      docker.Transport.Name(),
		destinationTransport: docker.Transport.Name(),
		scoped:               false,
	})

	if err != nil {
		return fmt.Errorf("syncing images failed: %w", err)
	}

	return nil
}
