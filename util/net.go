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

package util

import (
	"fmt"
	"net"
)

// AddressURL returns the URL of an HTTP address.
func AddressURL(address string) (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return "", err
	}
	url := "http://"
	if addr.IP == nil {
		url += "localhost"
	} else {
		url += addr.IP.String()
	}
	if addr.Port != 0 {
		url += fmt.Sprintf(":%d", addr.Port)
	}
	return url, nil
}
