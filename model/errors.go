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
	"errors"
)

// Errors.
var (
	ErrNotFound      = errors.New("not found")
	ErrType          = errors.New("wrong type")
	ErrFirstNegative = errors.New("first cannot be negative")
	ErrLastNegative  = errors.New("last cannot be negative")
	ErrNotRunning    = errors.New("project isn't running")
	ErrNotStopped    = errors.New("project isn't stopped")
)
