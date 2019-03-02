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

package subscriptiongen

import (
	"go/types"

	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/vektah/gqlparser/ast"

	"groundcontrol/plugin/util"
)

// Plugin generates subscriptions.
type Plugin struct {
	filename string
	pkgname  string
}

// SubscriptionBuild contains data about the subscriptions to build.
type SubscriptionBuild struct {
	Stored  []*Stored
	Deleted []*Deleted
}

// Stored contains data about a stored subscription to build.
type Stored struct {
	Type     types.Type
	TypeName string
}

// Deleted contains data about a deleted subscription to build.
type Deleted struct {
	Type     types.Type
	TypeName string
}

// New creates a new subscription plugin.
func New(filename, pkgname string) *Plugin {
	return &Plugin{filename: filename, pkgname: pkgname}
}

// Name returns the name of the plugin.
func (p *Plugin) Name() string {
	return "subscriptiongen"
}

// MutateConfig adds subscriptions.
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

	build := &SubscriptionBuild{}

	subscriptions := schema.Types["Subscription"]
	if subscriptions == nil {
		return nil
	}

	for _, field := range subscriptions.Fields {
		if directive := field.Directives.ForName("stored"); directive != nil {
			if err := p.stored(cfg, schema, field, build, binder); err != nil {
				return nil
			}
			continue
		}

		if directive := field.Directives.ForName("deleted"); directive != nil {
			if err := p.deleted(cfg, schema, field, build, binder); err != nil {
				return nil
			}
			continue
		}
	}

	return templates.Render(templates.Options{
		PackageName:     p.pkgname,
		Filename:        p.filename,
		Data:            build,
		GeneratedHeader: true,
	})
}

func (p *Plugin) stored(
	cfg *config.Config,
	schema *ast.Schema,
	field *ast.FieldDefinition,
	build *SubscriptionBuild,
	binder *config.Binder,
) error {
	gqlType := field.Type
	name := gqlType.Name()
	def := schema.Types[name]

	goType, err := util.GoType(cfg, schema, binder, name)
	if err != nil {
		return err
	}

	build.Stored = append(build.Stored, &Stored{
		Type:     util.CopyModifiersFromAst(gqlType, def, goType),
		TypeName: templates.ToGo(name),
	})

	return nil
}

func (p *Plugin) deleted(
	cfg *config.Config,
	schema *ast.Schema,
	field *ast.FieldDefinition,
	build *SubscriptionBuild,
	binder *config.Binder,
) error {
	gqlType := field.Type
	name := gqlType.Name()
	def := schema.Types[name]

	goType, err := util.GoType(cfg, schema, binder, name)
	if err != nil {
		return err
	}

	build.Deleted = append(build.Deleted, &Deleted{
		Type:     util.CopyModifiersFromAst(gqlType, def, goType),
		TypeName: templates.ToGo(name),
	})

	return nil
}
