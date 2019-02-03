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

import "context"

// System contains information about the running app.
type System struct {
	ID               string   `json:"id"`
	JobIDs           []string `json:"jobsIds"`
	JobMetricsID     string   `json:"jobMetricsId"`
	ProcessGroupIDs  []string `json:"processGroupIds"`
	ProcessMetricsID string   `json:"processMetricsId"`
	LogEntryIDs      []string `json:"logEntryIds"`
	LogMetricsID     string   `json:"logMetricsId"`
}

// IsNode tells gqlgen that it implements Node.
func (System) IsNode() {}

// Jobs returns paginated jobs.
func (s System) Jobs(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
	status []JobStatus,
) (JobConnection, error) {
	var slice []Job

	nodes := GetModelContext(ctx).Nodes

	for _, nodeID := range s.JobIDs {
		node := nodes.MustLoadJob(nodeID)
		match := len(status) == 0

		for _, v := range status {
			if node.Status == v {
				match = true
				break
			}
		}

		if match {
			slice = append(slice, node)
		}
	}

	return PaginateJobSlice(
		slice,
		after,
		before,
		first,
		last,
	)
}

// JobMetrics returns the JobMetrics node.
func (s System) JobMetrics(ctx context.Context) JobMetrics {
	return GetModelContext(ctx).Nodes.MustLoadJobMetrics(s.JobMetricsID)
}

// ProcessGroups returns paginated process groups.
func (s System) ProcessGroups(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
	status []ProcessStatus,
) (ProcessGroupConnection, error) {
	var slice []ProcessGroup

	nodes := GetModelContext(ctx).Nodes

	for _, nodeID := range s.ProcessGroupIDs {
		node := nodes.MustLoadProcessGroup(nodeID)
		match := len(status) == 0

		for _, v := range status {
			if node.Status(ctx) == v {
				match = true
				break
			}
		}

		if match {
			slice = append(slice, node)
		}
	}

	return PaginateProcessGroupSlice(
		slice,
		after,
		before,
		first,
		last,
	)
}

// ProcessMetrics returns the ProcessMetrics node.
func (s System) ProcessMetrics(ctx context.Context) ProcessMetrics {
	return GetModelContext(ctx).Nodes.MustLoadProcessMetrics(s.ProcessMetricsID)
}

// LogEntries returns paginated log entries.
func (s System) LogEntries(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
	level []LogLevel,
	ownerID *string,
) (LogEntryConnection, error) {
	var slice []LogEntry

	nodes := GetModelContext(ctx).Nodes

	for _, nodeID := range s.LogEntryIDs {
		node := nodes.MustLoadLogEntry(nodeID)

		if ownerID != nil && *ownerID != node.OwnerID {
			continue
		}

		match := len(level) == 0

		for _, v := range level {
			if node.Level == v {
				match = true
				break
			}
		}

		if match {
			slice = append(slice, node)
		}
	}

	return PaginateLogEntrySlice(
		slice,
		after,
		before,
		first,
		last,
	)
}

// LogMetrics returns the LogMetrics node.
func (s System) LogMetrics(ctx context.Context) LogMetrics {
	return GetModelContext(ctx).Nodes.MustLoadLogMetrics(s.LogMetricsID)
}
