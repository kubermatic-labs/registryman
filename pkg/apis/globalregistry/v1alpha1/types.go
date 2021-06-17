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
package v1alpha1

import (
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

//
// mmmmm                  "             m
// #   "#  mmm    mmmm  mmm     mmm   mm#mm   m mm  m   m
// #mmmm" #"  #  #" "#    #    #   "    #     #"  " "m m"
// #   "m #""""  #   #    #     """m    #     #      #m#
// #    " "#mm"  "#m"#  mm#mm  "mmm"    "mm   #      "#
//                m  #                               m"
//                 ""                               ""
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=registries,scope=Cluster,singular=registry
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Registry describes the expected state of a registry Object
type Registry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec describes the Registry Specification.
	Spec *RegistrySpec `json:"spec"`
}

// Registry implements the runtime.Object interface
var _ runtime.Object = &Registry{}

// RegistrySpec describes the specification of a Registry.
type RegistrySpec struct {

	// +kubebuilder:validation:Enum=harbor;acr

	// Provider identifies the actual registry type, e.g. Harbor, Docker Hub,
	// etc.
	Provider string `json:"provider"`

	// +kubebuilder:validation:Pattern=`^(https?|ftp)://[^\s/$.?#].[^\s]*$`

	// APIEndpoint identifies the registry API endpoint in a registry
	// implementation specific way. It can be for example an HTTP endpoint,
	// like "http://harbor.example.com:8080".
	APIEndpoint string `json:"apiEndpoint"`

	// Username is the user name to be used during the authentication at the
	// APIEndpoint interface.
	Username string `json:"username"`

	// Password is the password to be used during the authentication at the
	// APIEndpoint interface.
	Password string `json:"password"`

	// +kubebuilder:default=Local
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=GlobalHub;Local

	// Role specifies whether the registry is a Global Hub or a Local
	// registry.
	Role string `json:"role"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RegistryList collects Registry resources.
type RegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Registry `json:"items"`
}

//
// mmmmm                   "                   m
// #   "#  m mm   mmm    mmm    mmm    mmm   mm#mm
// #mmm#"  #"  " #" "#     #   #"  #  #"  "    #
// #       #     #   #     #   #""""  #        #
// #       #     "#m#"     #   "#mm"  "#mm"    "mm
//                         #
//                       ""
//

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Project describes the expected state of a globalregistry Project
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              *ProjectSpec `json:"spec"`
}

// Project implements the runtime.Object interface
var _ runtime.Object = &Project{}

// ProjectSpec describes the spec field of the Project resource
type ProjectSpec struct {
	// Type selects whether the project is global or local.
	Type ProjectType `json:"type"`

	// LocalRegistries lists the registry names at which the local project
	// shall be provisioned at.
	// +listType=set
	LocalRegistries []string `json:"localRegistries,omitempty"`
	// Members enumerates the project members and their capabilities
	// provisioned for the specific registry.
	// +listType=set
	Members []*ProjectMember `json:"members"`
}

//------------------------------------------------

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Enum=Global;Local

// ProjectType specifies whether a project is Global or Local.
type ProjectType int

const (
	// GlobalProjectType is a registry type that hosts all global
	// projects which are then replicated to all registries of the
	// LocalProjectType.
	GlobalProjectType ProjectType = iota
	// LocalProjectType is a registry type that hosts selected local
	// projects and all Global projects.
	LocalProjectType
)

func (rt ProjectType) string() (string, error) {
	switch rt {
	default:
		return "", fmt.Errorf("unhandled RegistryType (%d)", rt)
	case GlobalProjectType:
		return "Global", nil
	case LocalProjectType:
		return "Local", nil
	}
}

// String method implements the Stringer interface
func (rt ProjectType) String() string {
	s, err := rt.string()
	if err != nil {
		panic(err)
	}
	return s
}

// MarshalJSON implements the Marshaller interface
func (rt ProjectType) MarshalJSON() ([]byte, error) {
	s, err := rt.string()
	if err != nil {
		return nil, err
	}
	return json.Marshal(s)
}

func (rt *ProjectType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	switch s {
	default:
		return fmt.Errorf("failed unmarshalling %s to ProjectType", s)
	case "Global":
		*rt = GlobalProjectType
	case "Local":
		*rt = LocalProjectType
	}
	return nil
}

// ProjectMember reprensents a User, Group or Robot user of a Project.
type ProjectMember struct {
	// Type of the project member, e.g. User, Group, Robot. If not set, the
	// default value (User) is applied.
	Type MemberType `json:"type,omitempty"`
	// Name of the project member
	Name string `json:"name"`

	// Role of the project member, e.g. Developer, Maintainer, etc.
	//
	// The possible values depend on the value of the Type field.
	Role MemberRole `json:"role"`

	// +kubebuilder:validation:Optional
	// DN is optional distinguished name of the user. Used with LDAP integration.
	DN string `json:"dn,omitempty"`
}

func (pm *ProjectMember) UnmarshalJSON(data []byte) error {
	type innerProjectMember ProjectMember

	// Setting the default values
	defaultPM := &innerProjectMember{
		Type: UserMemberType,
	}
	if err := json.Unmarshal(data, defaultPM); err != nil {
		return err
	}
	*pm = ProjectMember(*defaultPM)
	return nil
}

// +kubebuilder:validation:Type=string

// MemberType selects the type of the membership, like, User, Group or Robot.
type MemberType int

const (
	UserMemberType MemberType = iota
	GroupMemberType
	RobotMemberType
)

func (mt MemberType) string() (string, error) {
	switch mt {
	default:
		return "", fmt.Errorf("unhandled MemberType (%d)", mt)
	case UserMemberType:
		return "User", nil
	case GroupMemberType:
		return "Group", nil
	case RobotMemberType:
		return "Robot", nil
	}
}

func (mt MemberType) String() string {
	s, err := mt.string()
	if err != nil {
		panic(err)
	}
	return s
}

func (mt MemberType) MarshalJSON() ([]byte, error) {
	s, err := mt.string()
	if err != nil {
		return nil, err
	}
	return json.Marshal(s)
}

func (mt *MemberType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	switch s {
	default:
		return fmt.Errorf("failed unmarshalling %s to MemberType", s)
	case "User":
		*mt = UserMemberType
	case "Group":
		*mt = GroupMemberType
	case "Robot":
		*mt = RobotMemberType
	}
	return nil
}

// +kubebuilder:validation:Type=string

type MemberRole int

const (
	LimitedGuestRole MemberRole = iota
	GuestRole
	DeveloperRole
	MaintainerRole
	ProjectAdminRole
	PushOnlyRole
	PullOnlyRole
	PullAndPushRole
)

func (mr MemberRole) string() (string, error) {
	switch mr {
	default:
		return "", fmt.Errorf("unhandled MemberRole (%d)", mr)
	case LimitedGuestRole:
		return "LimitedGuest", nil
	case GuestRole:
		return "Guest", nil
	case DeveloperRole:
		return "Developer", nil
	case MaintainerRole:
		return "Maintainer", nil
	case ProjectAdminRole:
		return "ProjectAdmin", nil
	case PushOnlyRole:
		return "PushOnly", nil
	case PullOnlyRole:
		return "PullOnly", nil
	case PullAndPushRole:
		return "PullAndPush", nil
	}
}

func (mr MemberRole) String() string {
	s, err := mr.string()
	if err != nil {
		panic(err)
	}
	return s
}

func (mr MemberRole) MarshalJSON() ([]byte, error) {
	s, err := mr.string()
	if err != nil {
		return nil, err
	}
	return json.Marshal(s)
}

func (mr *MemberRole) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	switch s {
	default:
		return fmt.Errorf("failed unmarshalling %s to MemberRole", s)
	case "LimitedGuest":
		*mr = LimitedGuestRole
	case "Guest":
		*mr = GuestRole
	case "Developer":
		*mr = DeveloperRole
	case "Maintainer":
		*mr = MaintainerRole
	case "ProjectAdmin":
		*mr = ProjectAdminRole
	case "PushOnly":
		*mr = PushOnlyRole
	case "PullOnly":
		*mr = PullOnlyRole
	case "PullAndPush":
		*mr = PullAndPushRole
	}
	return nil
}
