package main

import (
	"fmt"
	"github.com/dolmen-go/codegen"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func MapItems[K comparable, V interface{}](mapObj map[K]V) []V {
	var values []V
	for _, v := range mapObj {
		values = append(values, v)
	}

	return values
}

func Map[T1 interface{}, T2 interface{}](arr *[]T1, f func(T1) T2) []T2 {
	var values []T2
	for _, v := range *arr {
		values = append(values, f(v))
	}

	return values
}

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

type Struct struct {
	Name      string
	IsClass   bool
	Super     *string
	Vtable    *string
	Functions []Function
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
		retTypes := Map(&d.Type.Results.List, typeToString)
		paramTypes := Map(&d.Type.Params.List, typeToString)
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

	structs = MapItems(structByName)
	return
}

func Write(path string, packageName string, structs []*Struct) error {
	const template = `// Code generated by example_test.go; DO NOT EDIT.
package {{.packageName}}

{{range $struct := .structs}}

{{if ne $struct.Vtable nil}}
type {{$struct.Vtable}} struct {
{{range $func := $struct.Functions}}
	{{$func.Name}} func({{$func.ArgTypesString}}) ({{$func.RetTypesString}})
{{end}}
}

func (this *{{$struct.Name}}) initVtable() {
{{range $func := $struct.Functions}}
	this.{{$func.Name}} = this.{{$func.Name}}Impl
{{end}}
}
{{end}}

{{if ne $struct.Super nil}}
func (this *{{$struct.Name}}) super() *{{$struct.Super}} {
	return &this.{{$struct.Super}}
}

func (this *{{$struct.Name}}) initVtable() {
{{range $func := $struct.Functions}}
	this.{{$func.Name}} = this.{{$func.Name}}Impl
{{end}}
}
{{end}}

{{end}}
`

	var filteredStructs []*Struct
	for _, s := range structs {
		if !s.IsClass {
			continue
		}

		// Copy the given struct keeping only methods ending in 'Impl' while removing it from the name.
		cls := NewStruct(s.Name)
		cls.Super = s.Super
		cls.Vtable = s.Vtable
		cls.IsClass = true
		for _, f := range s.Functions {
			if !strings.HasSuffix(f.Name, "Impl") {
				continue
			}

			cls.Functions = append(cls.Functions, Function{
				strings.TrimSuffix(f.Name, "Impl"),
				f.ArgumentTypes,
				f.ReturnTypes,
			})
		}

		filteredStructs = append(filteredStructs, cls)
	}

	tmpl := codegen.MustParse(template)
	if err := tmpl.CreateFile(path, map[string]interface{}{
		"structs":     filteredStructs,
		"packageName": packageName,
	}); err != nil {
		return err
	}
	log.Printf("File %s created.\n", path)

	return nil
}

func main() {
	fmt.Printf("Gooping..")
	fmt.Printf("%v\n", os.Args)
	for _, path := range os.Args[1:] {
		if !strings.HasSuffix(path, ".go") {
			path += ".go"
		}

		dir, file := filepath.Split(strings.TrimSuffix(path, ".go"))
		outputPath := filepath.Join(dir, fmt.Sprintf("%s_goop.go", file))

		packageName, structs, err := Parse(path)
		if err != nil {
			panic(err)
		}

		fmt.Printf("\n\nPackage: %v\n", packageName)
		for _, s := range structs {
			fmt.Printf("%v\n", s)
		}

		err = Write(outputPath, packageName, structs)
		if err != nil {
			panic(err)
		}
	}
}
