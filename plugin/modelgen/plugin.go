package modelgen

import (
	"errors"
	"go/types"
	"sort"
	"strings"

	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/99designs/gqlgen/plugin"
	"github.com/vektah/gqlparser/ast"

	"groundcontrol/plugin/util"
)

// ModelBuild contains data about the models to build.
type ModelBuild struct {
	PackageName string
	Interfaces  []*Interface
	Models      []*Object
	Enums       []*Enum
	Connections []*Connection
}

// Interface contains data about an interface to build.
type Interface struct {
	Description string
	Name        string
}

// Object contains data about an object to build.
type Object struct {
	Description string
	Name        string
	Fields      []*Field
	Implements  []string
	Relates     []*Relate
	Paginates   []*Paginate
}

// Field contains data about an object field to build.
type Field struct {
	Description string
	Name        string
	Type        types.Type
	Tag         string
	Slice       bool
}

// Enum contains data about an enum to build.
type Enum struct {
	Description string
	Name        string
	Raw         string
	Values      []*EnumValue
}

// EnumValue contains data about an enum value to build.
type EnumValue struct {
	Description string
	Name        string
	Value       string
}

// Relate contains data about a relation to build.
type Relate struct {
	Description   string
	Name          string
	TypeName      string
	Type          types.Type
	NonNull       bool
	GoIDFieldName string
}

// Paginate contains data about a pagination to build.
type Paginate struct {
	Description    string
	Name           string
	NodeName       string
	Connection     types.Type
	Edge           types.Type
	Node           types.Type
	GoIDsFieldName string
	Filters        []*Argument
}

// Argument contains data about an argument to build.
type Argument struct {
	Name string
	Type types.Type
}

// Connection contains data about a connection to build.
type Connection struct {
	Name     string
	NodeName string
	EdgeName string
	Edge     types.Type
	Node     types.Type
}

// New creates a new plugin to build model.
func New() plugin.Plugin {
	return &Plugin{}
}

// Plugin represents the plugin to build model.
type Plugin struct{}

var _ plugin.ConfigMutator = &Plugin{}

// Name returns the name of the plugin.
func (m *Plugin) Name() string {
	return "modelgen"
}

// MutateConfig adds model.
func (m *Plugin) MutateConfig(cfg *config.Config) error {
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

	b := &ModelBuild{
		PackageName: cfg.Model.Package,
	}

	for _, schemaType := range schema.Types {
		if cfg.Models.UserDefined(schemaType.Name) {
			continue
		}

		switch schemaType.Kind {
		case ast.Interface, ast.Union:
			it := &Interface{
				Description: schemaType.Description,
				Name:        templates.ToGo(schemaType.Name),
			}

			b.Interfaces = append(b.Interfaces, it)
		case ast.Object, ast.InputObject:
			if schemaType == schema.Query || schemaType == schema.Mutation || schemaType == schema.Subscription {
				continue
			}
			it := &Object{
				Description: schemaType.Description,
				Name:        templates.ToGo(schemaType.Name),
			}

			for _, implementor := range schema.GetImplements(schemaType) {
				it.Implements = append(it.Implements, templates.ToGo(implementor.Name))
			}

			if strings.HasSuffix(schemaType.Name, "Connection") {
				if err := m.connection(cfg, schema, schemaType, b, binder); err != nil {
					return err
				}
			}

			for _, field := range schemaType.Fields {
				if directive := field.Directives.ForName("dynamic"); directive != nil {
					continue
				}

				typ, err := util.GoType(cfg, schema, binder, field.Type.Name())
				if err != nil {
					return err
				}

				name := field.Name
				if nameOveride := cfg.Models[schemaType.Name].Fields[field.Name].FieldName; nameOveride != "" {
					name = nameOveride
				}

				if directive := field.Directives.ForName("relate"); directive != nil {
					if err := m.relate(cfg, schema, field, directive, it, binder); err != nil {
						return err
					}
					continue
				}

				if directive := field.Directives.ForName("paginate"); directive != nil {
					if err := m.paginate(cfg, schema, schemaType, field, directive, it, binder); err != nil {
						return err
					}
					continue
				}

				fd := schema.Types[field.Type.Name()]
				it.Fields = append(it.Fields, &Field{
					Name:        templates.ToGo(name),
					Type:        util.CopyModifiersFromAst(field.Type, fd, typ),
					Description: field.Description,
					Tag:         `json:"` + field.Name + `"`,
					Slice:       field.Type.Elem != nil,
				})
			}

			sort.Slice(it.Relates, func(i, j int) bool { return it.Relates[i].Name < it.Relates[j].Name })
			sort.Slice(it.Paginates, func(i, j int) bool { return it.Paginates[i].Name < it.Paginates[j].Name })

			b.Models = append(b.Models, it)
		case ast.Enum:
			it := &Enum{
				Name:        templates.ToGo(schemaType.Name),
				Raw:         schemaType.Name,
				Description: schemaType.Description,
			}

			for _, v := range schemaType.EnumValues {
				it.Values = append(it.Values, &EnumValue{
					Name:        templates.ToGo(v.Name),
					Value:       v.Name,
					Description: v.Description,
				})
			}

			b.Enums = append(b.Enums, it)
		}
	}

	sort.Slice(b.Enums, func(i, j int) bool { return b.Enums[i].Name < b.Enums[j].Name })
	sort.Slice(b.Models, func(i, j int) bool { return b.Models[i].Name < b.Models[j].Name })
	sort.Slice(b.Interfaces, func(i, j int) bool { return b.Interfaces[i].Name < b.Interfaces[j].Name })
	sort.Slice(b.Connections, func(i, j int) bool { return b.Connections[i].Name < b.Connections[j].Name })

	for _, it := range b.Enums {
		cfg.Models.Add(it.Raw, cfg.Model.ImportPath()+"."+it.Name)
	}
	for _, it := range b.Models {
		cfg.Models.Add(it.Name, cfg.Model.ImportPath()+"."+it.Name)
	}
	for _, it := range b.Interfaces {
		cfg.Models.Add(it.Name, cfg.Model.ImportPath()+"."+it.Name)
	}

	if len(b.Models) == 0 && len(b.Enums) == 0 {
		return nil
	}

	return templates.Render(templates.Options{
		PackageName:     cfg.Model.Package,
		Filename:        cfg.Model.Filename,
		Data:            b,
		GeneratedHeader: true,
	})
}

func (m *Plugin) connection(
	cfg *config.Config,
	schema *ast.Schema,
	schemaType *ast.Definition,
	build *ModelBuild,
	binder *config.Binder,
) error {
	connectionName := schemaType.Name
	connectionDef := schema.Types[connectionName]

	edgesField := connectionDef.Fields.ForName("edges")
	if edgesField == nil {
		return errors.New("edges must be an array in " + connectionName)
	}

	edgeType := edgesField.Type.Elem
	if edgeType == nil {
		return errors.New("edges must be an array in " + connectionName)
	}

	edgeName := edgeType.Name()
	edgeDef := schema.Types[edgeName]

	nodeField := edgeDef.Fields.ForName("node")
	if nodeField == nil {
		return errors.New("node must be a field in " + edgeName)
	}

	nodeType := nodeField.Type
	nodeName := nodeType.Name()
	nodeDef := schema.Types[nodeName]

	edgeGoType, err := util.GoType(cfg, schema, binder, edgeName)
	if err != nil {
		return err
	}

	nodeGoType, err := util.GoType(cfg, schema, binder, nodeName)
	if err != nil {
		return err
	}

	build.Connections = append(build.Connections, &Connection{
		Name:     templates.ToGo(connectionName),
		NodeName: templates.ToGo(nodeName),
		EdgeName: templates.ToGo(edgeName),
		Edge:     util.CopyModifiersFromAst(edgeType, edgeDef, edgeGoType),
		Node:     util.CopyModifiersFromAst(nodeType, nodeDef, nodeGoType),
	})

	return nil
}

func (m *Plugin) relate(
	cfg *config.Config,
	schema *ast.Schema,
	field *ast.FieldDefinition,
	directive *ast.Directive,
	obj *Object,
	binder *config.Binder,
) error {
	name := field.Name
	idName := templates.ToGo(name) + "ID"

	if idArgs := directive.Arguments.ForName("goIdFieldName"); idArgs != nil {
		idName = idArgs.Value.Raw
	}

	obj.Fields = append(obj.Fields, &Field{
		Name:        idName,
		Type:        types.Typ[types.String],
		Description: idName + " contains the ID of the related " + name + ".",
		Tag:         `json:"` + idName + `"`,
	})

	relateGoType, err := util.GoType(cfg, schema, binder, field.Type.Name())
	if err != nil {
		return err
	}

	fd := schema.Types[field.Type.Name()]

	obj.Relates = append(obj.Relates, &Relate{
		Description:   field.Description,
		Name:          templates.ToGo(name),
		TypeName:      templates.ToGo(field.Type.Name()),
		Type:          util.CopyModifiersFromAst(field.Type, fd, relateGoType),
		NonNull:       field.Type.NonNull,
		GoIDFieldName: idName,
	})

	return nil
}

func (m *Plugin) paginate(
	cfg *config.Config,
	schema *ast.Schema,
	schemaType *ast.Definition,
	field *ast.FieldDefinition,
	directive *ast.Directive,
	obj *Object,
	binder *config.Binder,
) error {
	name := field.Name
	idsName := templates.ToGo(name) + "IDs"

	if idsArgs := directive.Arguments.ForName("goIdsFieldName"); idsArgs != nil {
		idsName = idsArgs.Value.Raw
	}

	obj.Fields = append(obj.Fields, &Field{
		Name:        idsName,
		Type:        types.NewSlice(types.Typ[types.String]),
		Description: idsName + " contains the IDs of all the " + name + " related to the " + schemaType.Name + ".",
		Tag:         `json:"` + idsName + `"`,
		Slice:       true,
	})

	connectionName := field.Type.Name()
	connectionDef := schema.Types[connectionName]
	connectionType := field.Type

	edgesField := connectionDef.Fields.ForName("edges")
	if edgesField == nil {
		return errors.New("edges must be an array in " + connectionName)
	}

	edgeType := edgesField.Type.Elem
	if edgeType == nil {
		return errors.New("edges must be an array in " + connectionName)
	}

	edgeName := edgeType.Name()
	edgeDef := schema.Types[edgeName]

	nodeField := edgeDef.Fields.ForName("node")
	if nodeField == nil {
		return errors.New("node must be a field in " + edgeName)
	}

	nodeType := nodeField.Type
	nodeName := nodeType.Name()
	nodeDef := schema.Types[nodeName]

	connectionGoType, err := util.GoType(cfg, schema, binder, connectionName)
	if err != nil {
		return err
	}

	edgeGoType, err := util.GoType(cfg, schema, binder, edgeName)
	if err != nil {
		return err
	}

	nodeGoType, err := util.GoType(cfg, schema, binder, nodeName)
	if err != nil {
		return err
	}

	paginates := &Paginate{
		Name:           templates.ToGo(name),
		NodeName:       templates.ToGo(nodeName),
		Connection:     util.CopyModifiersFromAst(connectionType, connectionDef, connectionGoType),
		Edge:           util.CopyModifiersFromAst(edgeType, edgeDef, edgeGoType),
		Node:           util.CopyModifiersFromAst(nodeType, nodeDef, nodeGoType),
		Description:    field.Description,
		GoIDsFieldName: idsName,
	}

	for _, argument := range field.Arguments[4:] {
		argumentGoType, err := util.GoType(cfg, schema, binder, argument.Type.Name())
		if err != nil {
			return err
		}

		argumentDef := schema.Types[argument.Type.Name()]

		paginates.Filters = append(paginates.Filters, &Argument{
			Name: argument.Name,
			Type: util.CopyModifiersFromAst(argument.Type, argumentDef, argumentGoType),
		})
	}

	obj.Paginates = append(obj.Paginates, paginates)

	return nil
}
