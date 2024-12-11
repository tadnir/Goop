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

		//fmt.Printf("%v %T\n", decl, decl)
		//d, ok := decl.(*ast.GenDecl)
		//if !ok || d.Tok != token.TYPE {
		//	fmt.Printf("HERERE %+v\n", decl)
		//	continue
		//}
		//
		//if len(d.Specs) != 1 {
		//	return nil, fmt.Errorf("many specs %v", d.Specs)
		//}
		//
		//cls, ok := d.Specs[0].(*ast.TypeSpec)
		//if !ok {
		//	fmt.Printf("Nothing to do with spec %T\n", d.Specs[0])
		//	continue
		//}

		//currStruct := NewStruct(cls.Name.Name)
		//for _, field := range cls.Type.(*ast.StructType).Fields.List {
		//	if field.Tag == nil {
		//		continue
		//	}
		//
		//	//tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
		//	//tagValue, ok := tag.Lookup("goop")
		//	//if !ok {
		//	//	continue
		//	//}
		//	//
		//	//switch tagValue {
		//	//case "super":
		//	//	currStruct.IsClass = true
		//	//	currStruct.Super = &field.Type.(*ast.Ident).Name
		//	//case "vtable":
		//	//	currStruct.IsClass = true
		//	//	currStruct.Vtable = &field.Type.(*ast.Ident).Name
		//	//}
		//
		//}
		//
		//file.structs[currStruct.Name] = currStruct
	}

	//for _, decl := range f.Decls {
	//	d, ok := decl.(*ast.FuncDecl)
	//	// Only Receiver functions are currently supported(or relevant)
	//	if !ok || d.Recv == nil {
	//		continue
	//	}
	//
	//	//receiver := d.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
	//	name := d.Name.Name
	//	typeToString := func(f *ast.Field) string {
	//		switch fRetType := f.Type.(type) {
	//		case *ast.Ident:
	//			return fRetType.Name
	//		case *ast.StarExpr:
	//			return fRetType.X.(*ast.Ident).Name
	//		default:
	//			panic(fmt.Errorf("Unknown type %T %v", f.Type, f.Type))
	//		}
	//	}
	//	retTypes := utils.Map(slices.Values(d.Type.Results.List), typeToString)
	//	paramTypes := utils.Map(slices.Values(d.Type.Params.List), typeToString)
	//	//fmt.Printf("func (%v) %v(%v) (%v)\n", receiver, name, paramTypes, retTypes)
	//
	//	//structByName[receiver].Functions = append(structByName[receiver].Functions,
	//	fmt.Printf("%v",
	//		Function{
	//			Name:          name,
	//			ArgumentTypes: paramTypes,
	//			ReturnTypes:   retTypes,
	//		})
	//
	//	//// find the @class decorator
	//	//isClass = false
	//	//for _, comment := range tdecl.Doc.List {
	//	//	if strings.Contains(comment.Text, "@class") {
	//	//		isClass = true
	//	//		break
	//	//	}
	//	//}
	//	//if !isClass {
	//	//	continue
	//	//}
	//	//
	//	//class := Class{}
	//	//
	//	//// get the Name of the class
	//	//for _, spec := range tdecl.Specs {
	//	//	if ts, ok := spec.(*ast.TypeSpec); ok {
	//	//		if ts.Name == nil {
	//	//			continue
	//	//		}
	//	//		class.Name = ts.Name.Name
	//	//		break
	//	//	}
	//	//}
	//	//if class.Name == "" {
	//	//	return fmt.Errorf("Unable to extract Name from a class struct.")
	//	//}
	//	//
	//	//// parse the goop tag and build columns
	//	//sdecl := tdecl.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType)
	//	//fields := sdecl.Fields.List
	//	//for _, field := range fields {
	//	//
	//	//	if field.Tag == nil {
	//	//		continue
	//	//	}
	//	//
	//	//	match := tagPattern.FindStringSubmatch(field.Tag.Value)
	//	//	if len(match) == 2 {
	//	//
	//	//		//col := Column{}
	//	//		//if err := col.Init(field.Names[0].Name, match[1]); err != nil {
	//	//		//	return fmt.Errorf(
	//	//		//		"Unable to parse tag '%s' from table '%s' in '%s': %v",
	//	//		//		match[1],
	//	//		//		table.Name,
	//	//		//		path,
	//	//		//		err,
	//	//		//	)
	//	//		//}
	//	//		//table.Columns = append(table.Columns, col)
	//	//		//if col.IsPrimary {
	//	//		//	table.PrimaryKeys = append(table.PrimaryKeys, col)
	//	//		//}
	//	//	}
	//	//}
	//	//if len(table.Columns) > 0 && len(table.PrimaryKeys) > 0 {
	//	//	(*i).Tables = append((*i).Tables, table)
	//	//}
	//}

	return file, nil
}

func (file *GoFile) GetStructs() []*StructDeclaration {
	return slices.Collect(maps.Values(file.structs))
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
