package go_generator

import (
	"fmt"
	"github.com/tadnir/goop/utils"
	"slices"
	"strings"
)

type GoFunctionBuilder struct {
	name     string
	params   []goVarDecl
	retVals  []goVarDecl
	receiver *goFuncReceiver
	impl     string
}

type goFuncReceiver struct {
	name         string
	ReceiverType string
	isRef        bool
}

func NewGoFunctionBuilder(name string) *GoFunctionBuilder {
	if name == "" {
		panic("Structs must be given names")
	}

	return &GoFunctionBuilder{
		name:     name,
		params:   []goVarDecl{},
		retVals:  []goVarDecl{},
		receiver: nil,
		impl:     "",
	}
}

func (b *GoFunctionBuilder) AddParam(name string, paramType string) *GoFunctionBuilder {
	b.params = append(b.params, goVarDecl{name, paramType})
	return b
}

func (b *GoFunctionBuilder) AddReturnType(name string, retType string) *GoFunctionBuilder {
	if name == "" {
		panic(fmt.Sprintf("function %v return type must have a name", b.name))
	}

	b.retVals = append(b.retVals, goVarDecl{name, retType})
	return b
}

func (b *GoFunctionBuilder) SetReceiver(name string, receiverType string, isRef bool) *GoFunctionBuilder {
	b.receiver = &goFuncReceiver{
		name:         name,
		ReceiverType: receiverType,
		isRef:        isRef,
	}
	return b
}

func (b *GoFunctionBuilder) SetImplRaw(impl string) *GoFunctionBuilder {
	b.impl = impl + "\n"
	return b
}

func (b *GoFunctionBuilder) AddImplLines(impl ...string) *GoFunctionBuilder {
	b.impl += "\t" + strings.Join(impl, "\n\t") + "\n"
	return b
}

func (b *GoFunctionBuilder) Build() string {
	receiver := ""
	if b.receiver != nil {
		ref := ""
		if b.receiver.isRef {
			ref = "*"
		}
		receiver = fmt.Sprintf("(%v %v%v) ", b.receiver.name, ref, b.receiver.ReceiverType)
	}

	parameters := strings.Join(utils.Map(slices.Values(b.params), goVarDecl.String), ", ")
	retVals := strings.Join(utils.Map(slices.Values(b.retVals), goVarDecl.String), ", ")
	if len(b.retVals) > 0 {
		retVals = fmt.Sprintf("(%v)", retVals)
	}
	if retVals != "" {
		retVals += " "
	}

	return fmt.Sprintf("func %v%v(%v) %v{\n%v}\n", receiver, b.name, parameters, retVals, b.impl)
}
