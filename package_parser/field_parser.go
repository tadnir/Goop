package package_parser

import (
	"fmt"
	"go/ast"
)

type FieldDeclaration struct {
	Name    *string
	VarType string
}

func ParseFieldDeclaration(decl *ast.Field) *FieldDeclaration {
	var name *string = nil
	if len(decl.Names) > 1 {
		panic(fmt.Sprintf("unexpected number of names: %+v", decl.Names))
	} else if len(decl.Names) == 1 {
		name = &decl.Names[0].Name
	}

	switch expr := decl.Type.(type) {
	case *ast.Ident:
		return &FieldDeclaration{Name: name, VarType: expr.String()}
	case *ast.StarExpr:
		return &FieldDeclaration{Name: name, VarType: "*" + expr.X.(*ast.Ident).String()}
	default:
		panic(fmt.Sprintf("unknown type %T", expr))
	}
}

func (f FieldDeclaration) String() string {
	if f.Name != nil {
		return fmt.Sprintf("%s %s", *f.Name, f.VarType)
	}

	return fmt.Sprintf("%s", f.VarType)
}
