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

//+build ignore

package main

import (
	"fmt"
	"os"

	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"

	"groundcontrol/plugin/jobgen"
	"groundcontrol/plugin/modelgen"
	"groundcontrol/plugin/subscriptiongen"
)

func main() {
	cfg, err := config.LoadConfigFromDefaultLocations()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}

	opts := []api.Option{
		api.NoPlugins(),
		api.AddPlugin(modelgen.New()),
		api.AddPlugin(jobgen.New("resolvers/auto_jobs.go", "resolvers")),
		api.AddPlugin(subscriptiongen.New("resolvers/auto_subscriptions.go", "resolvers")),
	}

	if err = api.Generate(cfg, opts...); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(3)
	}
}
