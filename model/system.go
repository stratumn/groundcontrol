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

package model

import (
	"context"
	"encoding/base64"
	"fmt"
)

func (n *System) filterJobsNode(ctx context.Context, node *Job, status []JobStatus) bool {
	match := len(status) == 0

	for _, v := range status {
		if node.Status == v {
			match = true
			break
		}
	}

	return match
}

func (n *System) filterProcessGroupsNode(ctx context.Context, node *ProcessGroup, status []ProcessStatus) bool {
	match := len(status) == 0

	for _, v := range status {
		if node.Status(ctx) == v {
			match = true
			break
		}
	}

	return match
}

func (n *System) filterLogEntriesNode(ctx context.Context, node *LogEntry, level []LogLevel, ownerID *string) bool {
	if ownerID != nil && *ownerID != node.OwnerID {
		return false
	}

	match := len(level) == 0

	for _, v := range level {
		if node.Level == v {
			match = true
			break
		}
	}

	return match
}

// LastMessageID is the ID of the last PubSub message and can be used to not miss any message when subscribing.
func (n *System) LastMessageID(ctx context.Context) string {
	modelCtx := GetModelContext(ctx)
	lastMessageID := modelCtx.Subs.LastMessageID()
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprint(lastMessageID)))
}
