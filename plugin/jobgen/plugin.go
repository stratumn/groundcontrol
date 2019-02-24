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

package jobgen

import (
	"go/types"

	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/vektah/gqlparser/ast"

	"groundcontrol/plugin/util"
)

// Plugin generates mutations for job.
type Plugin struct {
	filename string
	pkgname  string
}

// JobBuild contains data about the subscriptions to build.
type JobBuild struct {
	JobType types.Type
	JobName string
	Jobs    []*Job
}

// Job contains data about a job subscription to build.
type Job struct {
	Name  string
	Multi bool
}

// New creates a new subscription plugin.
func New(filename, pkgname string) *Plugin {
	return &Plugin{filename: filename, pkgname: pkgname}
}

// Name returns the name of the plugin.
func (p *Plugin) Name() string {
	return "jobgen"
}

// MutateConfig adds job mutations.
func (p *Plugin) MutateConfig(cfg *config.Config) error {
	if err := cfg.Check(); err != nil {
		return err
	}

	schema, _, err := cfg.LoadSchema()
	if err != nil {
		return err
	}

	cfg.InjectBuiltins(schema)

	binder, err := cfg.NewBinder(schema)
	if err != nil {
		return err
	}

	jobType, err := util.GoType(cfg, schema, binder, "Job")
	if err != nil {
		return err
	}

	build := &JobBuild{
		JobType: types.NewPointer(jobType),
		JobName: "model.Job",
	}

	mutations := schema.Types["Mutation"]
	if mutations == nil {
		return nil
	}

	for _, field := range mutations.Fields {
		if directive := field.Directives.ForName("job"); directive != nil {
			if err := p.job(cfg, schema, field, build, binder); err != nil {
				return nil
			}
		}
	}

	return templates.Render(templates.Options{
		PackageName:     p.pkgname,
		Filename:        p.filename,
		Data:            build,
		GeneratedHeader: true,
	})
}

func (p *Plugin) job(
	cfg *config.Config,
	schema *ast.Schema,
	field *ast.FieldDefinition,
	build *JobBuild,
	binder *config.Binder,
) error {
	build.Jobs = append(build.Jobs, &Job{
		Name:  templates.ToGo(field.Name),
		Multi: field.Type.Elem != nil,
	})

	return nil
}
