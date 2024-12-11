package main

import (
	"fmt"
	"github.com/tadnir/goop/go_generator"
	"github.com/tadnir/goop/package_parser"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Example data structure for the implement function
//var (
//	vtable = VTable{
//		name: "aVtable",
//		functions: []VFunc{
//			{
//				name:      "getName",
//				signature: "func() string",
//			},
//		},
//	}
//	a = Class{
//		name:      "A",
//		super:     nil,
//		vtable:    &vtable,
//		overrides: nil,
//	}
//	b = Class{
//		name:   "B",
//		super:  &a,
//		vtable: nil,
//		overrides: []*Override{
//			{
//				overriddenVtable: &vtable,
//				functions: []VFunc{
//					{
//						name:      "getName",
//						signature: "func() string",
//					},
//				},
//			},
//		},
//	}
//	c = Class{
//		name:   "C",
//		super:  &b,
//		vtable: nil,
//		overrides: []*Override{
//			{
//				overriddenVtable: &vtable,
//				functions: []VFunc{
//					{
//						name:      "getName",
//						signature: "func() string",
//					},
//				},
//			},
//		},
//	}
//)

func ImplementClass(file *go_generator.GoFileBuilder, class *Class) error {
	if class.super != nil {
		file.AddFunction(go_generator.NewGoFunctionBuilder("super").
			SetReceiver("this", class.name, true).
			AddReturnType("super", "*"+class.super.name).
			AddImplLines(
				"this.initClass()",
				"return &this."+class.super.name,
			),
		)
	}

	if class.vtable != nil {
		vtableStruct := go_generator.NewGoStructBuilder(class.vtable.name)
		vtableStruct.AddVar(class.vtable.IsInitName(), "bool")
		for _, function := range class.vtable.functions {
			vtableStruct.AddVar(function.name, function.signature)
		}
		file.AddStruct(vtableStruct)
	}

	initFunc := go_generator.NewGoFunctionBuilder("initClass").
		SetReceiver("this", class.name, true)
	initExitConditions := []string{}
	if class.vtable != nil {
		initExitConditions = append(initExitConditions, "this."+class.vtable.IsInitName())
	}
	if class.overrides != nil {
		for _, override := range class.overrides {
			initExitConditions = append(initExitConditions, "this."+override.overriddenVtable.IsInitName())
		}
	}
	if len(initExitConditions) > 0 {
		initFunc.AddImplLines(
			fmt.Sprintf("if %v {", strings.Join(initExitConditions, " && ")),
			"return",
			"}",
			"",
		)
	}

	if class.super != nil {
		initFunc.AddImplLines(fmt.Sprintf("(&this.%v).initClass()", class.super.name), "")
	}

	if class.vtable != nil {
		initFunc.AddImplLines(
			fmt.Sprintf("// Initializing VTable '%v'", class.vtable.name),
			fmt.Sprintf("this.%v = true", class.vtable.IsInitName()))
		for _, function := range class.vtable.functions {
			initFunc.AddImplLines(fmt.Sprintf("this.%v = this.%vImpl", function.name, function.name))
		}
	}

	if class.overrides != nil {
		for _, override := range class.overrides {
			initFunc.AddImplLines(fmt.Sprintf("// Initializing Overrides for VTable '%v'", override.overriddenVtable.name))
			for _, function := range override.functions {
				initFunc.AddImplLines(fmt.Sprintf("this.%v = this.%vImpl", function.name, function.name))
			}
		}
	}

	file.AddFunction(initFunc)

	return nil
}

func getParameters() (fileName string, packageName string, packagePath string) {
	fileName = os.Getenv("GOFILE")
	if fileName == "" {
		log.Fatal("Empty GOFILE")
	}

	if !strings.HasSuffix(fileName, ".go") {
		log.Fatal("GOFILE must end with .go")
	}

	packageName = os.Getenv("GOPACKAGE")
	if packageName == "" {
		log.Fatal("Empty GOPACKAGE")
	}

	packagePath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	return
}

func main() {
	inputFile, packageName, packagePath := getParameters()
	fmt.Printf("Gooping...\n")

	packageData, err := package_parser.ParsePackage(packageName, packagePath, true)
	if err != nil {
		panic(err)
	}

	classes := NewClassesContainer()
	for _, st := range packageData.GetStructs() {
		for _, field := range st.Variables {
			goopTag, isGoop := field.Tag.Lookup("goop")
			if !isGoop {
				continue
			}

			class := classes.GetClass(st.Name)
			switch goopTag {
			case "super":
				fmt.Printf("%s is child of %s!\n", st.Name, field.VarType)
				class.super = classes.GetClass(field.VarType)
			case "vtable":
				fmt.Printf("%s has a vtable named %s!\n", st.Name, field.VarType)
				class.vtable = &VTable{name: field.VarType, functions: []VFunc{}}
			default:
				fmt.Printf("Unknown goop tag '%s'!\n", goopTag)
			}
		}
	}

	for _, cl := range classes.GetClassesSorted() {
		for _, recvFunc := range packageData.GetReceiverFunctions(cl.name) {
			if IsVirtualMethod(recvFunc) {
				// check if any of the parents has this function in it's vtable, if so create an override
				// if not, if there's a vtable for the struct add it to there
				// otherwise panic
				if vtable := cl.ChooseVTable(recvFunc); vtable != nil {
					fmt.Printf("Overriden %s for %s\n", recvFunc.Name, cl.name)
					cl.RegisterVirtual(recvFunc, vtable)
				} else {
					panic(fmt.Sprintf("Can't find vtable for %s in %s", recvFunc.Name, cl.name))
				}
			}
		}
	}

	fmt.Printf("%+v", classes)

	fileData, err := packageData.GetFile(inputFile)
	if err != nil {
		panic(err)
	}

	file := go_generator.NewGoFileBuilder("goop", packageData.GetName())
	for _, st := range fileData.GetStructs() {
		fmt.Printf("Implementing class %s...\n", st.Name)
		err = ImplementClass(file, classes.GetClass(st.Name))
		if err != nil {
			panic(err)
		}
	}

	source, err := file.Build()
	if err != nil {
		panic(err)
	}

	println(source)

	outputFile := inputFile[:len(inputFile)-len(".go")] + "_goop.go"
	err = os.WriteFile(filepath.Join(packagePath, outputFile), []byte(source), 0777)
	if err != nil {
		panic(err)
	}
}
