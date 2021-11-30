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

//+genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories="registryman"
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=registries,scope=Namespaced,singular=registry
// +kubebuilder:subresource:status

// Registry describes the expected state of a registry Object
type Registry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec describes the Registry Specification.
	Spec   *RegistrySpec   `json:"spec"`
	Status *RegistryStatus `json:"status,omitempty"`
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

// RegistryStatus specifies the status of a registry.
type RegistryStatus struct {
	// +listType=map
	// +listMapKey=name
	Projects     []ProjectStatus      `json:"projects"`
	Capabilities RegistryCapabilities `json:"capabilities"`
}

type RegistryCapabilities struct {
	// CanCreateProject shows whether the registry can create projects.
	CanCreateProject bool `json:"canCreateProject"`

	// CanDeleteProject shows whether the registry can delete projects.
	CanDeleteProject bool `json:"canDeleteProject"`

	// CanPullReplicate shows whether the registry can pull repositories
	// from remote registries.
	CanPullReplicate bool `json:"canPullReplicate"`

	// CanPushReplicate shows whether the registry can push repositories
	// from remote registries.
	CanPushReplicate bool `json:"canPushReplicate"`

	// CanManipulateProjectMembers shows whether the registry can add/remove
	// members to the projects.
	CanManipulateProjectMembers bool `json:"canManipulateProjectMembers"`

	// CanManipulateProjectScanners shows whether the registry can add/remove
	// scanners to the projects.
	CanManipulateProjectScanners bool `json:"canManipulateScanners"`

	// CanManipulateProjectReplicationRules shows whether the registry can
	// add/remove replication rules to the projects.
	CanManipulateProjectReplicationRules bool `json:"canManipulateReplicationRules"`

	// HasProjectMembers shows whether the registry understands the concept
	// of project membership.
	HasProjectMembers bool `json:"hasProjectMembers"`

	// HasProjectScanners shows whether the registry understands the concept
	// of project level vulnerability scanners.
	HasProjectScanners bool `json:"hasProjectScanners"`

	// HasProjectReplicationRules shows whether the registry understands the
	// concept of project level replication rules.
	HasProjectReplicationRules bool `json:"hasProjectReplicationRules"`

	// HasProjectStorageReport shows whether the registry understands the concept
	// of project level storage reporting.
	HasProjectStorageReport bool `json:"hasProjectStorageReport"`
}

// ProjectStatus specifies the status of a registry project.
type ProjectStatus struct {

	// Name of the project.
	Name string `json:"name"`

	// Members of the project.
	//
	// +listType=map
	// +listMapKey=name
	Members []MemberStatus `json:"members"`

	// Replication rules of the project.
	//
	// +listType=atomic
	ReplicationRules []ReplicationRuleStatus `json:"replicationRules"`

	// Storage used by the project in bytes.
	StorageUsed int `json:"storageUsed"`

	// Scanner of the project.
	ScannerStatus ScannerStatus `json:"scannerStatus"`
}

// MemberStatus specifies the status of a project member.
type MemberStatus struct {

	// Name of the project member.
	Name string `json:"name"`

	// Type of the project membership, like user, group, robot.
	Type string `json:"type"`

	// Role of the project member, like admin, developer, maintainer, etc.
	Role string `json:"role"`

	// Distinguished name of the project member. Empty when omitted.
	DN string `json:"dn,omitempty"`
}

// ReplicationRuleStatus specifies the status of project replication rule.
type ReplicationRuleStatus struct {

	// RemoteRegistryName indicates the name of the remote registry which
	// the current registry shall synchronize with.
	RemoteRegistryName string `json:"remoteRegistryName"`

	// Trigger describes the event that shall trigger the replication.
	Trigger ReplicationTrigger `json:"trigger"`

	// Direction shows whether the replication is of type pull or push.
	Direction string `json:"direction"`
}

// ScannerStatus specifies the status of a project's external vulnerability scanner.
type ScannerStatus struct {

	// Name of the scanner.
	Name string `json:"name"`

	// URL of the scanner.
	URL string `json:"url"`
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

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="registryman"
// +kubebuilder:resource:path=projects,scope=Namespaced,singular=project

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
	//
	// +listType=set
	// +kubebuilder:validation:Optional
	LocalRegistries []string `json:"localRegistries,omitempty"`

	// Members enumerates the project members and their capabilities
	// provisioned for the specific registry.
	//
	// +kubebuilder:validation:Optional
	// +listType=map
	// +listMapKey=name
	Members []*ProjectMember `json:"members,omitempty"`

	// +kubebuilder:validation:Optional

	// Scanner specifies the name of the assigned scanner.
	Scanner string `json:"scanner,omitempty"`

	// +kubebuilder:validation:Optional

	// Trigger specifies the preferred replication trigger. If it is not
	// possible to implement the selected replication trigger, the trigger
	// may be overridden.
	Trigger ReplicationTrigger `json:"trigger,omitempty"`
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

// MarshalText implements the Marshaller interface
func (rt ProjectType) MarshalText() ([]byte, error) {
	return []byte(rt.String()), nil
}

// UnmarshalText implements the Unmarshaller interface
func (rt *ProjectType) UnmarshalText(text []byte) error {
	switch string(text) {
	default:
		return fmt.Errorf("failed unmarshalling %s to ProjectType", string(text))
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
// +kubebuilder:validation:Enum=User;Group;Robot

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

// MarshalText implements the Marshaller interface
func (mt MemberType) MarshalText() ([]byte, error) {
	return []byte(mt.String()), nil
}

// UnmarshalText implements the Unmarshaller interface
func (mt *MemberType) UnmarshalText(text []byte) error {
	switch string(text) {
	default:
		return fmt.Errorf("failed unmarshalling %s to MemberType", string(text))
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

// MemberRole shows the capabilities, the role of the member within the project.
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

// MarshalText implements the Marshaller interface
func (mr MemberRole) MarshalText() ([]byte, error) {
	return []byte(mr.String()), nil
}

// UnmarshalText implements the Unmarshaller interface
func (mr *MemberRole) UnmarshalText(text []byte) error {
	switch string(text) {
	default:
		return fmt.Errorf("failed unmarshalling %s to MemberRole", string(text))
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

// +kubebuilder:validation:Type=string
type ReplicationTriggerType int

const (
	UndefinedRepliationTriggerType ReplicationTriggerType = iota
	ManualReplicationTriggerType
	EventBasedReplicationTriggerType
	CronReplicationTriggerType
)

func (rtt ReplicationTriggerType) String() string {
	switch rtt {
	case ManualReplicationTriggerType:
		return "manual"
	case EventBasedReplicationTriggerType:
		return "event_based"
	case CronReplicationTriggerType:
		return "cron"
	default:
		return "undefined"
	}
}

// MarshalText implements the Marshaller interface
func (rtt ReplicationTriggerType) MarshalText() ([]byte, error) {
	return []byte(rtt.String()), nil
}

// UnmarshalText implements the Unmarshaller interface
func (rtt *ReplicationTriggerType) UnmarshalText(text []byte) error {
	switch string(text) {
	case "manual":
		*rtt = ManualReplicationTriggerType
	case "event_based":
		*rtt = EventBasedReplicationTriggerType
	case "cron":
		*rtt = CronReplicationTriggerType
	default:
		*rtt = UndefinedRepliationTriggerType
	}
	return nil
}

// ReplicationTrigger indicates the replication trigger of a project.
type ReplicationTrigger struct {
	Type ReplicationTriggerType `json:"type"`

	// +kubebuilder:validation:Optional
	Schedule string `json:"schedule,omitempty"`
}

func (rt ReplicationTrigger) TriggerType() ReplicationTriggerType {
	return rt.Type
}

func (rt ReplicationTrigger) TriggerSchedule() string {
	return rt.Schedule
}

func (rt ReplicationTrigger) String() string {
	rType, err := rt.Type.MarshalText()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s %s", string(rType), rt.Schedule)
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProjectList collects Registry resources.
type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Project `json:"items"`
}

//  ____
// / ___|  ___ __ _ _ __  _ __   ___ _ __
// \___ \ / __/ _` | '_ \| '_ \ / _ \ '__|
//  ___) | (_| (_| | | | | | | |  __/ |
// |____/ \___\__,_|_| |_|_| |_|\___|_|

// +genclient
// +kubebuilder:resource:path=scanners,scope=Namespaced,singular=scanner
// +kubebuilder:resource:categories="registryman"
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Scanner resource describes the configuration of an external vulnerability
// scanner.
type Scanner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec describes the Scanner Specification.
	Spec *ScannerSpec `json:"spec"`
}

type ScannerSpec struct {
	// +kubebuilder:validation:Pattern=`^(https?|ftp)://[^\s/$.?#].[^\s]*$`

	// A base URL of the scanner adapter.
	Url string `json:"url,omitempty"`

	// An optional value of the HTTP Authorization header sent with each
	// request to the Scanner Adapter API.
	AccessCredential string `json:"accessCredential,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ScannerList collects Registry resources.
type ScannerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Scanner `json:"items"`
}
