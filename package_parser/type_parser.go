package package_parser

import (
	"fmt"
	"github.com/tadnir/goop/utils"
	"go/ast"
	"slices"
	"strings"
)

type StructDeclaration struct {
	Name      string
	Doc       *string
	Variables []*FieldDeclaration
}

type InterfaceDeclaration struct {
	Name string
	Doc  *string
}

func ParseTypeDeclaration(decl *ast.GenDecl) (*StructDeclaration, *InterfaceDeclaration) {
	if len(decl.Specs) != 1 {
		panic(fmt.Errorf("expected only one type declaration specs got %+v", decl.Specs))
	}

	switch expr := decl.Specs[0].(type) {
	case *ast.TypeSpec:
		name := expr.Name.String()

		var doc *string = nil
		if decl.Doc.Text() != "" {
			d := decl.Doc.Text()
			doc = &d
		}

		if expr.TypeParams != nil {
			fmt.Printf("Type parameters for \"%s\" are not yet supported: %+v\n", name, expr.TypeParams)
		}

		switch typeDecl := expr.Type.(type) {
		case *ast.InterfaceType:
			return nil, &InterfaceDeclaration{
				Name: name,
				Doc:  doc,
			}
		case *ast.StructType:
			return &StructDeclaration{
				Name:      name,
				Doc:       doc,
				Variables: utils.Map(slices.Values(typeDecl.Fields.List), ParseFieldDeclaration),
			}, nil
		default:
			panic(fmt.Errorf("expected type declaration to be struct or interface got %T", expr.Type))
		}

	//case *ast.ValueSpec:
	//	fmt.Printf("Skipping %+v %T\n", expr, expr)
	//case *ast.ImportSpec:
	//	fmt.Printf("Skipping %+v\n", expr)
	default:
		panic(fmt.Errorf("unknown declaration spec type: %T", expr))
	}
}

func (s *StructDeclaration) String() string {
	var sb strings.Builder
	if s.Doc != nil {
		sb.WriteString("// ")
		sb.WriteString(*s.Doc)
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("type %v struct {", s.Name))
	if len(s.Variables) > 0 {
		sb.WriteString("\n")
	}
	for _, v := range s.Variables {
		sb.WriteString(fmt.Sprintf("\t%v\n", v))
	}
	sb.WriteString("}")
	return sb.String()
}

func (s *InterfaceDeclaration) String() string {
	var sb strings.Builder
	if s.Doc != nil {
		sb.WriteString("// ")
		sb.WriteString(*s.Doc)
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("type %v interface {", s.Name))
	sb.WriteString("}")
	return sb.String()
}
