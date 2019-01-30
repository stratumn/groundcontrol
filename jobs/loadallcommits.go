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

package jobs

import (
	"github.com/stratumn/groundcontrol/models"
	"github.com/stratumn/groundcontrol/pubsub"
)

// LoadAllCommits creates jobs to load the commits of every project.
// It doesn't return errors but will output a log message when errors happen.
func LoadAllCommits(
	nodes *models.NodeManager,
	log *models.Logger,
	jobManager *models.JobManager,
	subs *pubsub.PubSub,
	getProjectPath models.ProjectPathGetter,
	getProjectCachePath ProjectCachePathGetter,
	viewerID string,
) []string {
	viewer := nodes.MustLoadUser(viewerID)

	var jobIDs []string

	for _, workspace := range viewer.Workspaces(nodes) {
		for _, project := range workspace.Projects(nodes) {
			if project.IsLoadingCommits {
				continue
			}

			jobID, err := LoadCommits(
				nodes,
				jobManager,
				subs,
				getProjectPath,
				getProjectCachePath,
				project.ID,
				models.JobPriorityNormal,
			)

			if err != nil {
				log.Error("LoadCommits Failed", struct {
					Project models.Project
					Error   string
				}{
					project,
					err.Error(),
				})
				continue
			}

			jobIDs = append(jobIDs, jobID)
		}
	}

	return jobIDs
}
