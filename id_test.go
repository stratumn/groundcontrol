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

package groundcontrol

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeID(t *testing.T) {
	type args struct {
		identifiers []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		"one string",
		args{[]string{"User"}},
		"VXNlcg==",
	}, {
		"two string",
		args{[]string{"User", "0"}},
		"VXNlcjow",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeID(tt.args.identifiers...); got != tt.want {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}

func TestDecodeID(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{{
		"one string",
		args{"VXNlcg=="},
		[]string{"User"},
		false,
	}, {
		"two string",
		args{"VXNlcjow"},
		[]string{"User", "0"},
		false,
	}, {
		"invalid id",
		args{"*"},
		nil,
		true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeID(tt.args.id)
			if (err != nil) != tt.wantErr {
				assert.Equal(t, got, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}
