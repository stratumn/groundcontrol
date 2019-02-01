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

// LogEntry represents a log entry in the app.
type LogEntry struct {
	ID        string   `json:"id"`
	Level     LogLevel `json:"level"`
	CreatedAt DateTime `json:"createdAt"`
	Message   string   `json:"message"`
	OwnerID   string   `json:"ownerID"`
}

// IsNode tells gqlgen that it implements Node.
func (LogEntry) IsNode() {}

// Owner returns the node associated with the LogEntry.
func (l LogEntry) Owner(nodes *NodeManager) Node {
	if l.OwnerID == "" {
		return nil
	}

	return nodes.MustLoad(l.OwnerID)
}
