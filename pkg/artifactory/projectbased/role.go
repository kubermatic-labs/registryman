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

package projectbased

import (
	"encoding/json"
)

type role int

const (
	projectAdminRole role = 1
	developerRole    role = 2
	guestRole        role = 3
	maintainerRole   role = 4
)

func (r *role) UnmarshalJSON(b []byte) error {
	var i int
	if err := json.Unmarshal(b, &i); err != nil {
		return err
	}
	*r = role(i)
	return nil
}

func (r role) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(r))
}

// String method implements the Stringer interface for Role.
func (r role) String() string {
	switch r {
	case projectAdminRole:
		return "Project Admin"
	case developerRole:
		return "Developer"
	case guestRole:
		return "Viewer"
	case maintainerRole:
		return "Release Manager"
	default:
		return "*unknown-role*"
	}
}
