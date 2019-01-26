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

package resolvers

import (
	"context"

	"github.com/stratumn/groundcontrol/models"
)

type systemResolver struct {
	*Resolver
}

func (r *systemResolver) Jobs(
	ctx context.Context,
	obj *models.System,
	after *string,
	before *string,
	first *int,
	last *int,
	status []models.JobStatus,
) (models.JobConnection, error) {
	return obj.Jobs(r.Nodes, after, before, first, last, status)
}

func (r *systemResolver) JobMetrics(
	ctx context.Context,
	obj *models.System,
) (models.JobMetrics, error) {
	return obj.JobMetrics(r.Nodes), nil
}

func (r *systemResolver) ProcessGroups(
	ctx context.Context,
	obj *models.System,
	after *string,
	before *string,
	first *int,
	last *int,
) (models.ProcessGroupConnection, error) {
	return obj.ProcessGroups(r.Nodes, after, before, first, last)
}

func (r *systemResolver) ProcessGroupMetrics(
	ctx context.Context,
	obj *models.System,
) (models.ProcessGroupMetrics, error) {
	return obj.ProcessGroupMetrics(r.Nodes), nil
}

func (r *systemResolver) LogEntries(
	ctx context.Context,
	obj *models.System,
	after *string,
	before *string,
	first *int,
	last *int,
	level []models.LogLevel,
) (models.LogEntryConnection, error) {
	return obj.LogEntries(r.Nodes, after, before, first, last, level)
}

func (r *systemResolver) LogMetrics(
	ctx context.Context,
	obj *models.System,
) (models.LogMetrics, error) {
	return obj.LogMetrics(r.Nodes), nil
}
