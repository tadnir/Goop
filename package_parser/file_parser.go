package package_parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
)

type GoFile struct {
	fileName    string
	packageName string
	structs     map[string]*Struct
}

type Struct struct {
	Name      string
	IsClass   bool
	Super     *string
	Vtable    *string
	Functions []Function
}

func ParsePackageFile(packagePath string, fileName string) (*GoFile, error) {
	file := &GoFile{fileName: fileName, structs: map[string]*Struct{}}
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

	// Get declarations
	for _, decl := range f.Decls {
		d, ok := decl.(*ast.GenDecl)
		if !ok || d.Tok != token.TYPE {
			continue
		}

		if len(d.Specs) != 1 {
			return nil, fmt.Errorf("many specs %v", d.Specs)
		}

		cls, ok := d.Specs[0].(*ast.TypeSpec)
		if !ok {
			fmt.Printf("Nothing to do with spec %T\n", d.Specs[0])
			continue
		}

		currStruct := NewStruct(cls.Name.Name)
		for _, field := range cls.Type.(*ast.StructType).Fields.List {
			if field.Tag == nil {
				continue
			}

			//tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
			//tagValue, ok := tag.Lookup("goop")
			//if !ok {
			//	continue
			//}
			//
			//switch tagValue {
			//case "super":
			//	currStruct.IsClass = true
			//	currStruct.Super = &field.Type.(*ast.Ident).Name
			//case "vtable":
			//	currStruct.IsClass = true
			//	currStruct.Vtable = &field.Type.(*ast.Ident).Name
			//}

		}

		file.structs[currStruct.Name] = currStruct
	}

	return file, nil
}

func (file *GoFile) GetStructs() []*Struct {
	return MapItems(file.structs)
}

func (file *GoFile) String() string {
	out := "File: " + file.fileName + "\n"
	for _, st := range file.structs {
		out += st.String() + "\n"
	}

	return out
}
