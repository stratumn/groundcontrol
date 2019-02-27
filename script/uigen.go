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

// +build ignore

package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/shurcooL/vfsgen"
)

func checkError(err error) {
	if err != nil && !os.IsNotExist(err) {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	var fs http.FileSystem = http.Dir("../groundcontrol-ui/build")

	err := os.Remove("auto_ui.go")
	checkError(err)

	err = vfsgen.Generate(fs, vfsgen.Options{
		PackageName:  "main",
		BuildTags:    "release",
		VariableName: "embeddedUI",
		Filename:     "auto_ui.go",
	})
	checkError(err)
}
