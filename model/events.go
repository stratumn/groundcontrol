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

import "context"

// BeforeStorer runs a function before it is stored.
type BeforeStorer interface {
	BeforeStore(ctx context.Context)
}

// AfterStorer runs a function after it is stored.
type AfterStorer interface {
	AfterStore(ctx context.Context)
}

// BeforeDeleter runs a function before it is deleted.
type BeforeDeleter interface {
	BeforeDelete(ctx context.Context)
}

// AfterDeleter runs a function after it is deleted.
type AfterDeleter interface {
	AfterDelete(ctx context.Context)
}
