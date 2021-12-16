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

package pathbased

import (
	"fmt"
	"sort"
	"strings"
)

type role int

const (
	projectAdminRole role = 1
	developerRole    role = 2
	guestRole        role = 3
	maintainerRole   role = 4
)

// String method implements the Stringer interface for Role.
func (r role) String() string {
	switch r {
	case projectAdminRole:
		return "r,mxm,d,w,m,n"
	case developerRole:
		return "r,d,w,n"
	case guestRole:
		return "r"
	case maintainerRole:
		return "r,d,w,m,n"
	default:
		return "*unknown-role*"
	}
}

func roleFromList(s []string) string {
	sort.Strings(s)
	if contains(s, "r") {
		if contains(s, "d") && contains(s, "w") && contains(s, "n") {
			if contains(s, "m") {
				if contains(s, "mxm") {
					return "ProjectAdmin"
				}
				return "Maintainer"
			}
			return "Developer"
		}
		return "Guest"
	}
	fmt.Printf("%v\n", s)
	return "*unknown-role*"
}

func contains(s []string, searchterm string) bool {
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm
}

func roleFromString(s string) (role, error) {
	for _, r := range strings.Split(s, ",") {

		switch r {
		case "ProjectAdmin":
			return projectAdminRole, nil
		case "Developer":
			return developerRole, nil
		case "Guest":
			return guestRole, nil
		case "Maintainer":
			return maintainerRole, nil
		default:
			return role(-1), fmt.Errorf("unknown role: %s", s)
		}
	}
	return role(-1), fmt.Errorf("unknown role: %s", s)
}
