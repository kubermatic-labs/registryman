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

package harbor

import (
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type scanner struct {
	id        string
	registry  *registry
	name      string
	url       string
	isDefault bool
}

var _ globalregistry.Scanner = &scanner{}

func (s *scanner) Delete() error {
	return s.registry.deleteScanner(s.id)
}

func (s *scanner) GetName() string {
	return s.name
}

func (s *scanner) GetURL() string {
	return s.url
}
