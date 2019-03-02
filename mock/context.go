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

package mock

import (
	"testing"

	"github.com/golang/mock/gomock"

	"groundcontrol/appcontext"
	"groundcontrol/relay"
)

// Base Node IDs.
var (
	ViewerID = relay.EncodeID("User")
	SystemID = relay.EncodeID("System")
)

// AppContext embed an app context with mocked interfaces and exposes the mocked types.
type AppContext struct {
	*appcontext.Context
	MockNodes    *MockNodes
	MockJobs     *MockJobs
	MockServices *MockServices
	MockSubs     *MockSubs
	MockSources  *MockSources
	MockKeys     *MockKeys
}

// NewAppContext returns an app context with mocked interfaces.
func NewAppContext(t *testing.T, ctrl *gomock.Controller) AppContext {
	nodes := NewMockNodes(ctrl)
	log := NewMockLog(ctrl)
	jobs := NewMockJobs(ctrl)
	services := NewMockServices(ctrl)
	subs := NewMockSubs(ctrl)
	sources := NewMockSources(ctrl)
	keys := NewMockKeys(ctrl)
	return AppContext{
		Context: &appcontext.Context{
			Nodes:    nodes,
			Log:      log,
			Jobs:     jobs,
			Services: services,
			Subs:     subs,
			Sources:  sources,
			Keys:     keys,
			ViewerID: ViewerID,
			SystemID: SystemID,
		},
		MockNodes:    nodes,
		MockJobs:     jobs,
		MockServices: services,
		MockSubs:     subs,
		MockSources:  sources,
		MockKeys:     keys,
	}
}
