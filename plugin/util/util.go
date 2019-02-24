package util

import (
	"go/types"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/vektah/gqlparser/ast"
)

func PkgAndType(name string) (string, string) {
	parts := strings.Split(name, ".")
	if len(parts) == 1 {
		return "", name
	}

	return NormalizeVendor(strings.Join(parts[:len(parts)-1], ".")), parts[len(parts)-1]
}

var modsRegex = regexp.MustCompile(`^(\*|\[\])*`)

func NormalizeVendor(pkg string) string {
	modifiers := modsRegex.FindAllString(pkg, 1)[0]
	pkg = strings.TrimPrefix(pkg, modifiers)
	parts := strings.Split(pkg, "/vendor/")
	return modifiers + parts[len(parts)-1]
}

var invalidPackageNameChar = regexp.MustCompile(`[^\w]`)

func SanitizePackageName(pkg string) string {
	return invalidPackageNameChar.ReplaceAllLiteralString(filepath.Base(pkg), "_")
}

func GoType(cfg *config.Config, schema *ast.Schema, binder *config.Binder, name string) (types.Type, error) {
	if cfg.Models.UserDefined(name) {
		pkg, typeName := PkgAndType(cfg.Models[name].Model[0])
		return binder.FindType(pkg, typeName)
	}

	return types.NewNamed(types.NewTypeName(0, cfg.Model.Pkg(), templates.ToGo(name), nil), nil, nil), nil
}

func CopyModifiersFromAst(t *ast.Type, d *ast.Definition, base types.Type) types.Type {
	if t.Elem != nil {
		return types.NewSlice(CopyModifiersFromAst(t.Elem, d, base))
	}

	if usePtr(t, d) {
		return types.NewPointer(base)
	}

	return base
}

func usePtr(t *ast.Type, d *ast.Definition) bool {
	switch d.Kind {
	case ast.Object, ast.InputObject:
		return true
	case ast.Interface:
		return false
	}

	return !t.NonNull
}
