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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

// RobotV2
type robot struct {
	UpdateTime   time.Time         `json:"update_time,omitempty"`
	Description  string            `json:"description,omitempty"`
	Level        string            `json:"level,omitempty"`
	Editable     bool              `json:"editable,omitempty"`
	CreationTime time.Time         `json:"creation_time,omitempty"`
	ExpiresAt    int               `json:"expires_at,omitempty"`
	Name         string            `json:"name,omitempty"`
	Secret       string            `json:"secret,omitempty"`
	Disable      bool              `json:"disable,omitempty"`
	Duration     int               `json:"duration,omitempty"`
	Id           int               `json:"id,omitempty"`
	Permissions  []robotPermission `json:"permissions,omitempty"`
}

var _ globalregistry.ProjectMember = &robot{}

func (r *robot) GetName() string {
	return r.Name
}

func (r *robot) GetType() string {
	return robotType
}

func (r *robot) GetRole() string {
	canPull := false
	canPush := false
	for _, permission := range r.Permissions {
		// TODO: robot accounts covering multiple projects are not
		// supported yet
		for _, access := range permission.Access {
			if access.Action == "push" &&
				access.Resource == "repository" {
				canPush = true
			}
			if access.Action == "pull" &&
				access.Resource == "repository" {
				canPull = true
			}
		}
	}
	switch {
	case canPull && canPush:
		return "PullAndPush"
	case canPull && !canPush:
		return "PullOnly"
	case !canPull && canPush:
		return "PushOnly"
	default:
		panic("robot role not supported")
	}
}

type robotPermission struct {
	Access    []access `json:"access,omitempty"`
	Kind      string   `json:"kind,omitempty"`
	Namespace string   `json:"namespace,omitempty"`
}

type access struct {
	Action   string `json:"action,omitempty"`
	Resource string `json:"resource,omitempty"`
	Effect   string `json:"effect,omitempty"`
}

// RobotCreated
type robotCreated struct {
	Secret       string    `json:"secret,omitempty"`
	CreationTime time.Time `json:"creation_time,omitempty"`
	Id           int       `json:"id,omitempty"`
	ExpiresAt    int       `json:"expires_at,omitempty"`
	Name         string    `json:"name,omitempty"`
}

func (p *projectAPI) getRobotMembers(projectID int) ([]*robot, error) {
	url := *p.reg.parsedUrl
	url.Path = fmt.Sprintf("%s/%d/robots", path, projectID)
	p.reg.logger.V(1).Info("creating new request", "url", url.String())
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}
	p.reg.logger.V(1).Info("sending HTTP request", "req-uri", req.RequestURI)

	req.SetBasicAuth(p.reg.GetUsername(), p.reg.GetPassword())

	resp, err := p.reg.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	robotMembersResult := []*robot{}

	err = json.NewDecoder(resp.Body).Decode(&robotMembersResult)
	if err != nil {
		p.reg.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		p.reg.logger.Info(b.String())
		fmt.Printf("body: %+v\n", b.String())
	}
	p.reg.logger.V(1).Info("robots parsed", "result", robotMembersResult)
	return robotMembersResult, err
}

func (p *projectAPI) createProjectRobotMember(robotMember *robot) (*robotCreated, error) {
	url := *p.reg.parsedUrl
	url.Path = "/api/v2.0/robots"
	reqBodyBuf := bytes.NewBuffer(nil)
	err := json.NewEncoder(reqBodyBuf).Encode(robotMember)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url.String(), reqBodyBuf)
	if err != nil {
		return nil, err
	}

	req.Header["Content-Type"] = []string{"application/json"}
	req.SetBasicAuth(p.reg.GetUsername(), p.reg.GetPassword())

	resp, err := p.reg.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	robotResult := &robotCreated{}

	err = json.NewDecoder(resp.Body).Decode(robotResult)

	if err != nil {
		p.reg.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		p.reg.logger.Info(b.String())
	}
	return robotResult, err
}

func (p *projectAPI) deleteProjectRobotMember(projectID int, robotMemberID int) error {
	url := *p.reg.parsedUrl
	url.Path = fmt.Sprintf("%s/%d/robots/%d", path, projectID, robotMemberID)
	p.reg.logger.V(1).Info("creating new request", "url", url.String())
	req, err := http.NewRequest(http.MethodDelete, url.String(), nil)
	if err != nil {
		return err
	}
	p.reg.logger.V(1).Info("sending HTTP request")

	req.Header["Content-Type"] = []string{"application/json"}
	req.SetBasicAuth(p.reg.GetUsername(), p.reg.GetPassword())

	resp, err := p.reg.do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}
