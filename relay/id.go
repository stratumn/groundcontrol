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

package relay

import (
	"encoding/base64"
	"strings"
)

// EncodeID encodes a gobal ID.
func EncodeID(identifiers ...string) string {
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(identifiers, ":")))
}

// DecodeID decodes a global ID.
func DecodeID(id string) ([]string, error) {
	bytes, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(bytes), ":"), nil
}
