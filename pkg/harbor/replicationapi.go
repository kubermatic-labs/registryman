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
	"strconv"
	"strings"

	"time"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

const replicationPolicyPath = "/api/v2.0/replication/policies"

type replicationAPI struct {
	reg *registry
}

func newReplicationAPI(reg *registry) *replicationAPI {
	return &replicationAPI{
		reg: reg,
	}
}

func (r *replicationAPI) List() ([]globalregistry.ReplicationRule, error) {
	// FIX: thread unsafe handling of parsedUrl
	r.reg.parsedUrl.Path = replicationPolicyPath
	r.reg.logger.V(1).Info("creating new request", "parsedUrl", r.reg.parsedUrl.String())
	req, err := http.NewRequest(http.MethodGet, r.reg.parsedUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	r.reg.logger.V(1).Info("sending HTTP request", "req-uri", req.RequestURI)

	req.SetBasicAuth(r.reg.GetUsername(), r.reg.GetPassword())

	resp, err := r.reg.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	replicationsResult := []*replicationResponseBody{}

	err = json.NewDecoder(resp.Body).Decode(&replicationsResult)
	if err != nil {
		r.reg.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		r.reg.logger.Info(b.String())
		fmt.Printf("body: %+v\n", b.String())
	}
	replicationRules := make([]globalregistry.ReplicationRule, 0)
	for _, replResult := range replicationsResult {
		dir, err := replResult.direction()
		if err != nil {
			return nil, err
		}
		remote, err := replResult.remote()
		if err != nil {
			return nil, err
		}

		if len(replResult.Filters) >= 1 {
			replicationRules = append(replicationRules, &replicationRule{
				ID:          replResult.Id,
				api:         r,
				name:        replResult.Name,
				projectName: strings.TrimSuffix(replResult.Filters[0].Value, "/**"),
				Dir:         dir,
				ReplTrigger: replResult.Trigger,
				Remote:      remote,
			})
		}
	}

	return replicationRules, err
}

func (r *replicationAPI) create(project globalregistry.Project, remoteReg globalregistry.RegistryConfig, trigger globalregistry.ReplicationTrigger, direction globalregistry.ReplicationDirection) (globalregistry.ReplicationRule, error) {

	r.reg.logger.V(1).Info("ReplicationAPI.Create invoked",
		"project_name", project.GetName(),
		"remoteReg_name", remoteReg.GetName(),
		"trigger", trigger.String(),
		"direction", direction.String(),
	)
	local := &remoteRegistryStatus{
		Name:         "Local",
		CreationTime: time.Time{}.Format(time.RFC3339),
		Update_time:  time.Time{}.Format(time.RFC3339),
	}
	var replTrigger *replicationTrigger
	switch trigger {
	case globalregistry.ManualReplicationTrigger:
		replTrigger = &replicationTrigger{
			Type: "manual",
		}
	case globalregistry.EventReplicationTrigger:
		replTrigger = &replicationTrigger{
			Type: "event_based",
		}
	default:
		return nil, fmt.Errorf("invalid replication trigger: %d", trigger)
	}
	n := time.Now()
	now := n.Format(time.RFC3339)
	nowStamp := time.Now().Unix()
	replicationPolicy := &replicationResponseBody{
		CreationTime: now,
		UpdateTime:   now,
		Enabled:      true,
		Filters: []replicationFilter{
			{
				Type:  "name",
				Value: fmt.Sprintf("%s/**", project.GetName()),
			},
		},
		DestNamespace: "",
		Trigger:       replTrigger,
		Deletion:      true,
		Override:      true,
	}
	remoteRegistry, err := r.reg.remoteRegistries.getByNameOrCreate(remoteReg)
	if err != nil {
		return nil, err
	}
	var name string
	switch direction {
	case globalregistry.PushReplication:
		replicationPolicy.Description = fmt.Sprintf("Pushing %s project to %s on %s",
			project.GetName(),
			remoteReg.GetName(),
			replTrigger.Type,
		)
		replicationPolicy.SrcRegistry = local
		replicationPolicy.DestRegistry = remoteRegistry
		name = fmt.Sprintf("push-%s-to-%s-on-%s-%d",
			project.GetName(),
			remoteReg.GetName(),
			replTrigger.Type,
			nowStamp,
		)
		replicationPolicy.Name = name
	case globalregistry.PullReplication:
		replicationPolicy.Description = fmt.Sprintf("Pulling %s project from %s on %s",
			project.GetName(),
			remoteReg.GetName(),
			replTrigger.Type,
		)
		replicationPolicy.DestRegistry = local
		replicationPolicy.SrcRegistry = remoteRegistry
		name = fmt.Sprintf("pull-%s-from-%s-on-%s-%d",
			project.GetName(),
			remoteReg.GetName(),
			replTrigger.Type,
			nowStamp,
		)
		replicationPolicy.Name = name
	default:
		return nil, fmt.Errorf("unhandled replication direction: %d", direction)
	}

	reqBodyBuf := bytes.NewBuffer(nil)
	err = json.NewEncoder(reqBodyBuf).Encode(replicationPolicy)
	if err != nil {
		return nil, err
	}
	r.reg.logger.V(1).Info(reqBodyBuf.String())
	url := r.reg.parsedUrl
	url.Path = replicationPolicyPath
	req, err := http.NewRequest(http.MethodPost, url.String(), reqBodyBuf)
	if err != nil {
		return nil, err
	}

	req.Header["Content-Type"] = []string{"application/json"}
	req.SetBasicAuth(r.reg.GetUsername(), r.reg.GetPassword())

	resp, err := r.reg.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	replicationPolicyID, err := strconv.Atoi(strings.TrimPrefix(resp.Header.Get("Location"), replicationPolicyPath+"/"))
	if err != nil {
		r.reg.logger.Error(err, "cannot parse project ID from response Location header",
			"location-header", resp.Header.Get("Location"))
		return nil, err
	}

	return &replicationRule{
		ID:          replicationPolicyID,
		api:         r,
		name:        name,
		projectName: project.GetName(),
		Dir:         direction,
		ReplTrigger: replTrigger,
		Remote:      remoteRegistry,
	}, nil
}

func (r *replicationAPI) delete(id int) error {
	// FIX: thread unsafe handling of parsedUrl
	r.reg.parsedUrl.Path = fmt.Sprintf("%s/%d", replicationPolicyPath, id)
	r.reg.logger.V(1).Info("creating new request", "parsedUrl", r.reg.parsedUrl.String())
	req, err := http.NewRequest(http.MethodDelete, r.reg.parsedUrl.String(), nil)
	if err != nil {
		return err
	}
	r.reg.logger.V(1).Info("sending HTTP request", "req-uri", req.URL)

	req.Header["Content-Type"] = []string{"application/json"}
	// r.registry.AddBasicAuth(req)
	req.SetBasicAuth(r.reg.GetUsername(), r.reg.GetPassword())

	resp, err := r.reg.do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}
