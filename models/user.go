// Copyright 2019 Stratumn
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package models

import (
	"container/list"
	"io/ioutil"

	"github.com/stratumn/groundcontrol/relay"

	yaml "gopkg.in/yaml.v2"
)

// User contains all the data of the person using the app.
type User struct {
	ID         string      `json:"id"`
	Workspaces []Workspace `json:"workspaces"`
}

// Workspace returns the workspace with the given slug.
func (u *User) Workspace(slug string) *Workspace {
	for _, v := range u.Workspaces {
		if v.Slug == slug {
			return &v
		}
	}

	return nil
}

// Jobs returns paginated jobs and supports filtering by status.
func (u *User) Jobs(
	after *string,
	before *string,
	first *int,
	last *int,
	status []JobStatus,
) (*JobConnection, error) {
	jobList := list.New()
	element := GetJobList().Front()

	for element != nil {
		job := element.Value.(*Job)
		match := len(status) == 0

		for _, v := range status {
			if job.Status == v {
				match = true
				break
			}
		}

		if match {
			jobList.PushBack(job)
		}

		element = element.Next()
	}

	connection, err := jobPaginator.Paginate(jobList, after, before, first, last)
	if err != nil {
		return nil, err
	}

	edges := make([]JobEdge, len(connection.Edges))

	for i, v := range connection.Edges {
		edges[i] = JobEdge{
			Node:   v.Node.(*Job),
			Cursor: v.Cursor,
		}
	}

	return &JobConnection{
		Edges:    edges,
		PageInfo: connection.PageInfo,
	}, nil
}

// LoadUserYAML loads the content of a YAML file into a User model.
func LoadUserYAML(file string, user *User) error {
	user.ID = relay.EncodeID("User", file)

	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	if err := yaml.UnmarshalStrict(bytes, &user); err != nil {
		return err
	}

	for i := range user.Workspaces {
		workspace := &user.Workspaces[i]
		workspace.ID = relay.EncodeID("Workspace", workspace.Slug)

		for j := range workspace.Projects {
			project := &workspace.Projects[j]
			project.ID = relay.EncodeID(
				"Project",
				workspace.Slug,
				project.Repository,
				project.Branch,
			)
			project.Workspace = &user.Workspaces[i]
			project.commitList = list.New()
		}
	}

	return nil
}
