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

// +build tools

package main

// Import all the tools to include them in gomod.
import (
	_ "github.com/99designs/gqlgen/api"
	_ "github.com/99designs/gqlgen/codegen/config"
	_ "github.com/cortesi/modd/cmd/modd"
	_ "github.com/golang/mock/mockgen"
	_ "github.com/shurcooL/vfsgen"
)
