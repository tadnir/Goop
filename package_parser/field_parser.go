package package_parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"strings"
)

type FieldDeclaration struct {
	Name    *string
	VarType string
	Tag     reflect.StructTag
}

func ParseFieldDeclaration(decl *ast.Field) *FieldDeclaration {
	var name *string = nil
	if len(decl.Names) > 1 {
		panic(fmt.Sprintf("unexpected number of names: %+v", decl.Names))
	} else if len(decl.Names) == 1 {
		name = &decl.Names[0].Name
	}

	var tag reflect.StructTag
	if decl.Tag != nil {
		if decl.Tag.Kind != token.STRING {
			panic(fmt.Sprintf("unexpected tag type: %+v", decl.Tag))
		}
		tag = reflect.StructTag(strings.Trim(decl.Tag.Value, "`"))
	}

	switch expr := decl.Type.(type) {
	case *ast.Ident:
		return &FieldDeclaration{Name: name, VarType: expr.String(), Tag: tag}
	case *ast.StarExpr:
		return &FieldDeclaration{Name: name, VarType: "*" + expr.X.(*ast.Ident).String(), Tag: tag}
	default:
		panic(fmt.Sprintf("unknown field type %T", expr))
	}
}

func (f FieldDeclaration) String() string {
	if f.Name != nil {
		return fmt.Sprintf("%s %s", *f.Name, f.VarType)
	}

	return fmt.Sprintf("%s", f.VarType)
}
