package package_parser

import (
	"fmt"
	"github.com/tadnir/goop/utils"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"path/filepath"
	"slices"
	"strings"
)

type GoFile struct {
	fileName    string
	packageName string
	imports     []*Import
	functions   []*Function
	structs     map[string]*StructDeclaration
	interfaces  map[string]*InterfaceDeclaration
	variables   map[string]*FieldDeclaration
}

func ParseGoFile(packagePath string, fileName string) (*GoFile, error) {
	file := &GoFile{fileName: fileName, structs: map[string]*StructDeclaration{}, interfaces: map[string]*InterfaceDeclaration{}, variables: map[string]*FieldDeclaration{}}
	filePath := filepath.Join(packagePath, fileName)

	// Parse the file
	fileSet := token.NewFileSet()
	f, err := parser.ParseFile(fileSet, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse '%s': %s", filePath, err)
	}

	// Get the package name
	if f.Name == nil {
		return nil, fmt.Errorf("Missing package Name in '%s'", filePath)
	}
	file.packageName = f.Name.Name
	file.imports = utils.Map(slices.Values(f.Imports), ParseImport)

	// Get declarations
	for _, decl := range f.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			switch decl.Tok {
			case token.TYPE:
				stDecl, inDecl := ParseTypeDeclaration(decl)
				if stDecl != nil {
					file.structs[stDecl.Name] = stDecl
				}
				if inDecl != nil {
					file.interfaces[inDecl.Name] = inDecl
				}
			case token.VAR:
				fmt.Printf("var %v\n", decl)
			case token.CONST:
				fmt.Printf("const %v\n", decl)
			default:
				fmt.Printf("unknown tok: %v\n", decl.Tok)
			}
		case *ast.FuncDecl:
			function := ParseFunction(decl)
			file.functions = append(file.functions, function)
		}
	}

	return file, nil
}

func (file *GoFile) GetStructs() []*StructDeclaration {
	return slices.Collect(maps.Values(file.structs))
}

func (file *GoFile) GetFunctions() []*Function {
	return file.functions
}

func (file *GoFile) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("File: %v\n", file.fileName))
	if len(file.imports) > 0 {
		sb.WriteString("import (\n")
		for _, imp := range file.imports {
			sb.WriteString(fmt.Sprintf("\t%v\n", imp))
		}
		sb.WriteString(")\n")
	}

	for _, st := range file.structs {
		sb.WriteString(st.String())
		sb.WriteString("\n")
	}

	for _, in := range file.interfaces {
		sb.WriteString(in.String())
		sb.WriteString("\n")
	}

	for _, fn := range file.functions {
		sb.WriteString(fn.String())
		sb.WriteString("\n")
	}

	return sb.String()
}
