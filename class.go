package main

import (
	"fmt"
	"github.com/tadnir/goop/package_parser"
	"github.com/tadnir/goop/utils"
	"maps"
	"slices"
	"strings"
)

type ClassesContainer struct {
	classes map[string]*Class
}

func NewClassesContainer() *ClassesContainer {
	return &ClassesContainer{classes: make(map[string]*Class)}
}

func (c *ClassesContainer) GetClass(name string) *Class {
	if _, knownClass := c.classes[name]; !knownClass {
		c.classes[name] = &Class{name: name, overrides: []*Override{}}
	}
	return c.classes[name]
}

func (c *ClassesContainer) GetClassesSorted() []*Class {
	return slices.SortedStableFunc(maps.Values(c.classes), func(class *Class, class2 *Class) int {
		if class.super == nil && class2.super != nil {
			return -1
		}

		if class.super != nil && class2.super == nil {
			return 1
		}

		if class.super == class2 {
			return 1
		}

		if class2.super == class {
			return -1
		}

		return strings.Compare(class.name, class2.name)
	})
}

func (c *ClassesContainer) String() string {
	var sb strings.Builder
	for _, class := range c.GetClassesSorted() {
		sb.WriteString(class.String())
		sb.WriteString("\n")
	}

	return sb.String()
}

type Class struct {
	name      string
	super     *Class
	vtable    *VTable
	overrides []*Override
}

func (c *Class) String() string {
	var sb strings.Builder

	if c.super == nil {
		sb.WriteString(fmt.Sprintf("Class %s {\n", c.name))
	} else {
		sb.WriteString(fmt.Sprintf("Class %s : %s {\n", c.name, c.super.name))
	}

	if c.vtable != nil {
		sb.WriteString(fmt.Sprintf("VTable(%s):\n", c.vtable.name))
		for _, virt := range c.vtable.functions {
			sb.WriteString(fmt.Sprintf("\t%s\n", virt.signature))
		}
	}

	for _, override := range c.overrides {
		sb.WriteString(fmt.Sprintf("Overrides(%s):\n", override.overriddenVtable.name))
		for _, override := range override.overriddenVtable.functions {
			sb.WriteString(fmt.Sprintf("\t%s\n", override.signature))
		}
	}

	sb.WriteString("}\n")

	return sb.String()
}

func (c *Class) ChooseVTable(method *package_parser.Function) *VTable {
	for _, override := range c.overrides {
		if override.overriddenVtable.HasMethod(MethodVirtualName(method)) {
			return override.overriddenVtable
		}
	}

	if c.super != nil {
		superVTable := c.super.ChooseVTable(method)
		if superVTable != nil {
			return superVTable
		}
	}

	// May be nil
	return c.vtable
}

func (c *Class) HasVTable() bool {
	return c.vtable != nil
}

func (c *Class) RegisterVirtual(method *package_parser.Function, vtable *VTable) {
	if c.HasVTable() && c.vtable == vtable {
		c.vtable.AddVirtual(method)
		return
	}

	for _, override := range c.overrides {
		if override.overriddenVtable == vtable {
			override.AddOverride(method)
			return
		}
	}

	override := &Override{overriddenVtable: vtable}
	c.overrides = append(c.overrides, override)
	override.AddOverride(method)
}

type VTable struct {
	name      string
	functions []VFunc
}

func (v *VTable) HasMethod(methodName string) bool {
	for _, function := range v.functions {
		if function.name == methodName {
			return true
		}
	}
	return false
}

func (v *VTable) AddVirtual(method *package_parser.Function) {
	v.functions = append(v.functions, functionToVFunc(method))
}

func (v *VTable) IsInitName() string {
	return fmt.Sprintf("is%vInit", utils.Capitalize(v.name))
}

type VFunc struct {
	name      string
	signature string
}

type Override struct {
	overriddenVtable *VTable
	functions        []VFunc
}

func (o *Override) AddOverride(method *package_parser.Function) {
	o.functions = append(o.functions, functionToVFunc(method))
}

func functionToVFunc(method *package_parser.Function) VFunc {
	return VFunc{
		name:      MethodVirtualName(method),
		signature: method.Signature(),
	}
}

func IsVirtualMethod(method *package_parser.Function) bool {
	return strings.HasSuffix(method.Name, "Impl")
}

func MethodVirtualName(method *package_parser.Function) string {
	return strings.TrimSuffix(method.Name, "Impl")
}
