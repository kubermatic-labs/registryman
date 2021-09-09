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
	"context"
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

func (r *registry) getRobotMembers(ctx context.Context, projectID int) ([]*robot, error) {
	url := *r.parsedUrl
	url.Path = fmt.Sprintf("%s/%d/robots", path, projectID)
	r.logger.V(1).Info("creating new request", "url", url.String())
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}
	r.logger.V(1).Info("sending HTTP request", "req-uri", req.RequestURI)

	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(ctx, req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	robotMembersResult := []*robot{}

	err = json.NewDecoder(resp.Body).Decode(&robotMembersResult)
	if err != nil {
		r.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		r.logger.Info(b.String())
		fmt.Printf("body: %+v\n", b.String())
	}
	r.logger.V(1).Info("robots parsed", "result", robotMembersResult)
	return robotMembersResult, err
}

func (r *registry) createProjectRobotMember(ctx context.Context, robotMember *robot) (*robotCreated, error) {
	url := *r.parsedUrl
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
	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(ctx, req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	robotResult := &robotCreated{}

	err = json.NewDecoder(resp.Body).Decode(robotResult)

	if err != nil {
		r.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		r.logger.Info(b.String())
	}
	return robotResult, err
}

func (r *registry) deleteProjectRobotMember(ctx context.Context, projectID int, robotMemberID int) error {
	url := *r.parsedUrl
	url.Path = fmt.Sprintf("%s/%d/robots/%d", path, projectID, robotMemberID)
	r.logger.V(1).Info("creating new request", "url", url.String())
	req, err := http.NewRequest(http.MethodDelete, url.String(), nil)
	if err != nil {
		return err
	}
	r.logger.V(1).Info("sending HTTP request")

	req.Header["Content-Type"] = []string{"application/json"}
	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(ctx, req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}
