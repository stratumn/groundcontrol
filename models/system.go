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

// System contains information about the running app.
type System struct {
	ID               string   `json:"id"`
	JobIDs           []string `json:"jobsIDs"`
	JobMetricsID     string   `json:"jobMetricsID"`
	ProcessGroupIDs  []string `json:"processGroupIDs"`
	ProcessMetricsID string   `json:"processMetricsID"`
	LogEntryIDs      []string `json:"logEntryIDs"`
	LogMetricsID     string   `json:"logMetricsID"`
}

// IsNode tells gqlgen that it implements Node.
func (System) IsNode() {}

// Jobs returns paginated jobs.
func (s System) Jobs(
	nodes *NodeManager,
	after *string,
	before *string,
	first *int,
	last *int,
	status []JobStatus,
) (JobConnection, error) {
	var slice []Job

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
func (s System) JobMetrics(nodes *NodeManager) JobMetrics {
	return nodes.MustLoadJobMetrics(s.JobMetricsID)
}

// ProcessGroups returns paginated process groups.
func (s System) ProcessGroups(
	nodes *NodeManager,
	after *string,
	before *string,
	first *int,
	last *int,
	status []ProcessStatus,
) (ProcessGroupConnection, error) {
	var slice []ProcessGroup

	for _, nodeID := range s.ProcessGroupIDs {
		node := nodes.MustLoadProcessGroup(nodeID)
		match := len(status) == 0

		for _, v := range status {
			if node.Status(nodes) == v {
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
func (s System) ProcessMetrics(nodes *NodeManager) ProcessMetrics {
	return nodes.MustLoadProcessMetrics(s.ProcessMetricsID)
}

// LogEntries returns paginated log entries.
func (s System) LogEntries(
	nodes *NodeManager,
	after *string,
	before *string,
	first *int,
	last *int,
	level []LogLevel,
) (LogEntryConnection, error) {
	var slice []LogEntry

	for _, nodeID := range s.LogEntryIDs {
		node := nodes.MustLoadLogEntry(nodeID)
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
func (s System) LogMetrics(nodes *NodeManager) LogMetrics {
	return nodes.MustLoadLogMetrics(s.LogMetricsID)
}
