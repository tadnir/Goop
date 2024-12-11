package package_parser

import (
	"fmt"
	"go/ast"
)

type Import struct {
	alias *string
	path  string
}

func ParseImport(imp *ast.ImportSpec) *Import {
	if imp.Name != nil {
		return &Import{alias: &imp.Name.Name, path: imp.Path.Value}
	} else {
		return &Import{alias: nil, path: imp.Path.Value}
	}
}

func (i *Import) String() string {
	if i.alias != nil {
		return fmt.Sprintf("%v %v", *i.alias, i.path)
	} else {
		return fmt.Sprintf("%v", i.path)
	}
}
