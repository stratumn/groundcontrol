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

package shell

import (
	"context"
	"io"
	"strings"

	"mvdan.cc/sh/expand"
	"mvdan.cc/sh/interp"
	"mvdan.cc/sh/syntax"

	"groundcontrol/appcontext"
)

// Embedded runs shell commands using the mvdan.cc/sh package.
type Embedded struct {
	parser *syntax.Parser
	runner *interp.Runner
}

// NewEmbedded create a new Embedded shell.
func NewEmbedded(stdout, stderr io.Writer, dir string, env []string) (appcontext.Runner, error) {
	parser := syntax.NewParser()
	runner, err := interp.New(
		interp.StdIO(nil, stdout, stderr),
		interp.Dir(dir),
		interp.Env(expand.ListEnviron(env...)),
		interp.Module(interp.ModuleExec(execCmd)),
	)
	if err != nil {
		return nil, err
	}
	return &Embedded{parser: parser, runner: runner}, nil
}

func (s *Embedded) Run(ctx context.Context, command string) error {
	r := strings.NewReader(command)
	node, err := s.parser.Parse(r, "")
	if err != nil {
		return err
	}
	return s.runner.Run(ctx, node)
}
