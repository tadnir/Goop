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

type Class struct {
	name      string
	super     *Class
	vtable    *VTable
	overrides []*Override
}

type VTable struct {
	name      string
	functions []VFunc
}

type VFunc struct {
	name      string
	signature string
}

type Override struct {
	overriddenVtable string
	functions        []VFunc
}

func ImplementClass(file *go_generator.GoFileBuilder, class Class) error {
	if class.super != nil {
		file.AddFunction(go_generator.NewGoFunctionBuilder("Super").
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
		vtableStruct.AddVar(fmt.Sprintf("is%vInit", class.vtable.name), "bool")
		for _, function := range class.vtable.functions {
			vtableStruct.AddVar(function.name, function.signature)
		}
		file.AddStruct(vtableStruct)
	}

	initFunc := go_generator.NewGoFunctionBuilder("initClass").
		SetReceiver("this", class.name, true)
	initExitConditions := []string{}
	if class.vtable != nil {
		initExitConditions = append(initExitConditions, fmt.Sprintf("this.is%vInit", class.vtable.name))
	}
	if class.overrides != nil {
		for _, override := range class.overrides {
			initExitConditions = append(initExitConditions, fmt.Sprintf("this.is%vInit", override.overriddenVtable))
		}
	}
	if len(initExitConditions) > 0 {
		initFunc.AddImplLines(
			fmt.Sprintf("if %v {", strings.Join(initExitConditions, " && ")),
			"\treturn",
			"}",
			"",
		)
	}

	if class.super != nil {
		initFunc.AddImplLines(fmt.Sprintf("(&this.%v).initClass()", class.super.name))
	}

	if class.vtable != nil {
		initFunc.AddImplLines(fmt.Sprintf("this.is%vInit = true", class.vtable.name))
		for _, function := range class.vtable.functions {
			initFunc.AddImplLines(fmt.Sprintf("this.%v = this.%vImpl", function.name, function.name))
		}
	}

	if class.overrides != nil {
		for _, override := range class.overrides {
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

	for _, file := range packageData.GetFiles() {
		println(file.String())
	}

	outputFile := inputFile[:len(inputFile)-len(".go")] + "_goop.go"
	file := go_generator.NewGoFileBuilder("github.com/tadnir/goop", packageData.GetName())
	err = ImplementClass(file, Class{
		name: "C",
		super: &Class{
			name: "B",
			super: &Class{
				name:  "A",
				super: nil,
				vtable: &VTable{
					name: "AVtable",
					functions: []VFunc{
						{
							name:      "getName",
							signature: "func() string",
						},
					},
				},
				overrides: nil,
			},
			vtable: nil,
			overrides: []*Override{
				{
					overriddenVtable: "AVtable",
					functions: []VFunc{
						{
							name:      "getName",
							signature: "func() string",
						},
					},
				},
			},
		},
		vtable: nil,
		overrides: []*Override{
			{
				overriddenVtable: "AVtable",
				functions: []VFunc{
					{
						name:      "getName",
						signature: "func() string",
					},
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(filepath.Join(packagePath, outputFile), []byte(file.Build()), 0777)
	if err != nil {
		panic(err)
	}
}
