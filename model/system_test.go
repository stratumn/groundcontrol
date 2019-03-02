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
	"groundcontrol/appcontext"
	"groundcontrol/mock"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSystem_filterJobsNode(t *testing.T) {
	type args struct {
		node   *Job
		status []JobStatus
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		"filter nil Status",
		args{&Job{Status: JobStatusRunning}, nil},
		true,
	}, {
		"filter empty Status",
		args{&Job{Status: JobStatusRunning}, []JobStatus{}},
		true,
	}, {
		"filter one Status",
		args{&Job{Status: JobStatusRunning}, []JobStatus{JobStatusRunning}},
		true,
	}, {
		"filter two Statuses",
		args{&Job{Status: JobStatusRunning}, []JobStatus{JobStatusQueued, JobStatusRunning}},
		true,
	}, {
		"don't filter Status",
		args{&Job{Status: JobStatusRunning}, []JobStatus{JobStatusFailed}},
		false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &System{}
			got := n.filterJobsNode(context.Background(), tt.args.node, tt.args.status)
			assert.Equal(t, tt.want, got, "System.filterJobsNode()")
		})
	}
}

func TestSystem_filterLogEntriesNode(t *testing.T) {
	one := "1"
	type args struct {
		node    *LogEntry
		level   []LogLevel
		ownerID *string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		"filter nil status nil OwnerID",
		args{&LogEntry{Level: LogLevelWarning}, nil, nil},
		true,
	}, {
		"filter empty Level nil OwnerID",
		args{&LogEntry{Level: LogLevelWarning}, []LogLevel{}, nil},
		true,
	}, {
		"filter one Level nil OwnerID",
		args{&LogEntry{Level: LogLevelWarning}, []LogLevel{LogLevelWarning}, nil},
		true,
	}, {
		"filter two Leveles nil OwnerID",
		args{&LogEntry{Level: LogLevelWarning}, []LogLevel{LogLevelDebug, LogLevelWarning}, nil},
		true,
	}, {
		"don't filter Level nil OwnerID",
		args{&LogEntry{Level: LogLevelWarning}, []LogLevel{LogLevelInfo}, nil},
		false,
	}, {
		"filter nil Level OwnerID",
		args{&LogEntry{Level: LogLevelWarning, OwnerID: "1"}, nil, &one},
		true,
	}, {
		"don't filter nil Level OwnerID",
		args{&LogEntry{Level: LogLevelWarning, OwnerID: "2"}, nil, &one},
		false,
	}, {
		"filter one Level OwnerID",
		args{&LogEntry{Level: LogLevelWarning, OwnerID: "1"}, []LogLevel{LogLevelWarning}, &one},
		true,
	}, {
		"don't filter one Level OwnerID",
		args{&LogEntry{Level: LogLevelWarning, OwnerID: "2"}, []LogLevel{LogLevelWarning}, &one},
		false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &System{}
			got := n.filterLogEntriesNode(context.Background(), tt.args.node, tt.args.level, tt.args.ownerID)
			assert.Equal(t, tt.want, got, "System.filterLogEntriesNode()")
		})
	}
}

func TestSystem_LastMessageID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCtx := mock.NewAppContext(t, ctrl)
	ctx := appcontext.With(context.Background(), mockCtx.Context)
	system := System{}
	mockCtx.MockSubs.EXPECT().LastMessageID().Return(uint64(100))
	got := system.LastMessageID(ctx)
	assert.Equal(t, "MTAw", got, "System.LastMessageID()")
}
