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

//go:generate go run scripts/subscriptionsgen.go -t User,System,DirectorySource,GitSource,Workspace,Project,Task,Key,Job,ProcessGroup,Process,LogEntry,JobMetrics,ProcessMetrics,LogMetrics -o resolvers/auto_subscriptions.go
//go:generate go run scripts/gqlgen.go

package main

import (
	"log"
	"net/http"
	"os"

	"groundcontrol/cmd"
)

var ui http.FileSystem

func main() {
	log.SetOutput(os.Stdout)
	cmd.Execute(ui)
}
