package go_generator

import (
	"fmt"
	"strings"
)

type GoStructBuilder struct {
	name      string
	vars      []goVarDecl
	functions []*GoFunctionBuilder
}

func NewGoStructBuilder(name string) *GoStructBuilder {
	return &GoStructBuilder{name: name}
}

func (b *GoStructBuilder) AddVar(name string, varType string) *GoStructBuilder {
	b.vars = append(b.vars, goVarDecl{name, varType})
	return b
}

func (b *GoStructBuilder) AddReceiverFunction(function *GoFunctionBuilder, receiverName string, isRefReceiver bool) *GoStructBuilder {
	function.setReceiver(receiverName, b.name, isRefReceiver)
	b.functions = append(b.functions, function)
	return b
}

func (b *GoStructBuilder) Build() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("type %v struct {\n", b.name))
	for _, v := range b.vars {
		sb.WriteString(fmt.Sprintf("\t%v\n", v.String()))
	}
	sb.WriteString("}\n")

	for _, f := range b.functions {
		sb.WriteString("\n")
		sb.WriteString(f.Build())
	}

	return sb.String()
}
