package package_parser

import (
	"fmt"
	"github.com/tadnir/goop/utils"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"reflect"
	"slices"
	"strings"
)

type Function struct {
	Name          string
	ArgumentTypes []string
	ReturnTypes   []string
}

func (f *Function) ArgTypesString() string {
	return strings.Join(f.ArgumentTypes, ", ")
}

func (f *Function) RetTypesString() string {
	return strings.Join(f.ReturnTypes, ", ")
}

func (f *Function) String() string {
	return fmt.Sprintf("%s(%v)(%v)", f.Name,
		f.ArgTypesString(), f.RetTypesString())
}

func NewStruct(name string) *Struct {
	return &Struct{Name: name, IsClass: false, Super: nil, Vtable: nil, Functions: []Function{}}
}

func (s Struct) String() string {
	typeName := "Struct"
	if s.IsClass {
		typeName = "Class"
	}

	desc := fmt.Sprintf("%v %v", typeName, s.Name)
	if s.Super != nil {
		desc += ": " + *s.Super
	}

	desc += " {\n"
	if s.Vtable != nil {
		desc += "\t<vtable: " + *s.Vtable + ">\n"
	}

	for _, f := range s.Functions {
		desc += "\t" + f.String() + ";\n"
	}
	return desc + "}"
}

func Parse(path string) (packageName string, structs []*Struct, err error) {
	fileSet := token.NewFileSet()
	f, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
	if err != nil {
		err = fmt.Errorf("Unable to parse '%s': %s", path, err)
		return
	}

	if f.Name == nil {
		err = fmt.Errorf("Missing package Name in '%s'", path)
		return
	}
	packageName = f.Name.Name

	structByName := map[string]*Struct{}
	for _, decl := range f.Decls {
		d, ok := decl.(*ast.GenDecl)
		if !ok || d.Tok != token.TYPE {
			continue
		}

		if len(d.Specs) != 1 {
			err = fmt.Errorf("many specs %v", d.Specs)
			return
		}

		cls, ok := d.Specs[0].(*ast.TypeSpec)
		if !ok {
			fmt.Printf("Nothing to do with spec %T\n", d.Specs[0])
			continue
		}

		currStruct := NewStruct(cls.Name.Name)
		fmt.Printf("%v\n", cls.Name.Name)
		for _, field := range cls.Type.(*ast.StructType).Fields.List {
			if field.Tag == nil {
				continue
			}

			tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
			tagValue, ok := tag.Lookup("goop")
			if !ok {
				continue
			}

			switch tagValue {
			case "super":
				currStruct.IsClass = true
				currStruct.Super = &field.Type.(*ast.Ident).Name
			case "vtable":
				currStruct.IsClass = true
				currStruct.Vtable = &field.Type.(*ast.Ident).Name
			}

		}

		structByName[currStruct.Name] = currStruct
	}

	for _, decl := range f.Decls {
		d, ok := decl.(*ast.FuncDecl)
		// Only Receiver functions are currently supported(or relevant)
		if !ok || d.Recv == nil {
			continue
		}

		receiver := d.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
		name := d.Name.Name
		typeToString := func(f *ast.Field) string {
			switch fRetType := f.Type.(type) {
			case *ast.Ident:
				return fRetType.Name
			case *ast.StarExpr:
				return fRetType.X.(*ast.Ident).Name
			default:
				panic(fmt.Errorf("Unknown type %T %v", f.Type, f.Type))
			}
		}
		retTypes := utils.Map(slices.Values(d.Type.Results.List), typeToString)
		paramTypes := utils.Map(slices.Values(d.Type.Params.List), typeToString)
		//fmt.Printf("func (%v) %v(%v) (%v)\n", receiver, name, paramTypes, retTypes)

		structByName[receiver].Functions = append(structByName[receiver].Functions,
			Function{
				Name:          name,
				ArgumentTypes: paramTypes,
				ReturnTypes:   retTypes,
			})

		//// find the @class decorator
		//isClass = false
		//for _, comment := range tdecl.Doc.List {
		//	if strings.Contains(comment.Text, "@class") {
		//		isClass = true
		//		break
		//	}
		//}
		//if !isClass {
		//	continue
		//}
		//
		//class := Class{}
		//
		//// get the Name of the class
		//for _, spec := range tdecl.Specs {
		//	if ts, ok := spec.(*ast.TypeSpec); ok {
		//		if ts.Name == nil {
		//			continue
		//		}
		//		class.Name = ts.Name.Name
		//		break
		//	}
		//}
		//if class.Name == "" {
		//	return fmt.Errorf("Unable to extract Name from a class struct.")
		//}
		//
		//// parse the goop tag and build columns
		//sdecl := tdecl.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType)
		//fields := sdecl.Fields.List
		//for _, field := range fields {
		//
		//	if field.Tag == nil {
		//		continue
		//	}
		//
		//	match := tagPattern.FindStringSubmatch(field.Tag.Value)
		//	if len(match) == 2 {
		//
		//		//col := Column{}
		//		//if err := col.Init(field.Names[0].Name, match[1]); err != nil {
		//		//	return fmt.Errorf(
		//		//		"Unable to parse tag '%s' from table '%s' in '%s': %v",
		//		//		match[1],
		//		//		table.Name,
		//		//		path,
		//		//		err,
		//		//	)
		//		//}
		//		//table.Columns = append(table.Columns, col)
		//		//if col.IsPrimary {
		//		//	table.PrimaryKeys = append(table.PrimaryKeys, col)
		//		//}
		//	}
		//}
		//if len(table.Columns) > 0 && len(table.PrimaryKeys) > 0 {
		//	(*i).Tables = append((*i).Tables, table)
		//}
	}

	structs = slices.Collect(maps.Values(structByName))
	return
}
