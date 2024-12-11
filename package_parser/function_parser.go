package package_parser

import (
	"fmt"
	"github.com/tadnir/goop/utils"
	"go/ast"
	"slices"
	"strings"
)

type Function struct {
	Name          string
	Doc           *string
	ArgumentTypes []*FieldDeclaration
	ReturnTypes   []*FieldDeclaration
	Receiver      *FunctionReceiver
	Body          string
}

type FunctionReceiver struct {
	Name     *string
	RecvType string
	isRef    bool
}

func ParseFunction(decl *ast.FuncDecl) *Function {
	function := new(Function)

	function.Name = decl.Name.Name
	// TODO: Support function body parsing
	function.Body = "unsupported yet"
	if decl.Doc != nil {
		docString := strings.TrimSpace(decl.Doc.Text())
		function.Doc = &docString
	}

	if decl.Recv != nil {
		if len(decl.Recv.List) > 1 {
			panic(fmt.Sprintf("found multiple receivers for function \"%v\": %+v", function.Name, decl.Recv.List))
		}
		if len(decl.Recv.List[0].Names) != 1 {
			panic(fmt.Sprintf("expected one name for function's \"%v\" reciever, found: %+v", function.Name, decl.Recv.List[0].Names))
		}
		name := decl.Recv.List[0].Names[0].String()
		switch recvType := decl.Recv.List[0].Type.(type) {
		case *ast.StarExpr:
			recvTypeIdent, ok := recvType.X.(*ast.Ident)
			if !ok {
				panic(fmt.Sprintf("expected star receiver for function's \"%v\" type, found: %+v", function.Name, recvType))
			}
			function.Receiver = &FunctionReceiver{Name: &name, isRef: true, RecvType: recvTypeIdent.String()}
		case *ast.Ident:
			function.Receiver = &FunctionReceiver{Name: &name, isRef: false, RecvType: recvType.String()}
		default:
			fmt.Printf("unsupported receiver type: %T\n", decl.Recv.List[0].Type)
		}
	}

	if decl.Type.TypeParams != nil {
		fmt.Printf("type parameters are not supported, from function \"%v\": %+v", function.Name, decl.Type.TypeParams)
	}

	for _, arg := range decl.Type.Params.List {
		function.ArgumentTypes = append(function.ArgumentTypes, ParseFieldDeclaration(arg))
	}

	if decl.Type.Results != nil {
		for _, arg := range decl.Type.Results.List {
			function.ReturnTypes = append(function.ReturnTypes, ParseFieldDeclaration(arg))
		}
	}

	return function
}

func (f *Function) Declaration() string {
	var sb strings.Builder
	if f.Receiver != nil {
		sb.WriteString(fmt.Sprintf("(%s) ", f.Receiver))
	}

	sb.WriteString(fmt.Sprintf("%s(%v)", f.Name, strings.Join(utils.Map(slices.Values(f.ArgumentTypes), (*FieldDeclaration).String), ", ")))
	if f.ReturnTypes != nil {
		sb.WriteString(fmt.Sprintf(" (%s)", strings.Join(utils.Map(slices.Values(f.ReturnTypes), (*FieldDeclaration).String), ", ")))
	}

	return sb.String()
}

func (f *Function) Signature() string {
	parameters := strings.Join(utils.Map(slices.Values(f.ArgumentTypes), (*FieldDeclaration).String), ", ")
	returns := ""
	if f.ReturnTypes != nil {
		returns = fmt.Sprintf(" (%s)", strings.Join(utils.Map(slices.Values(f.ReturnTypes), (*FieldDeclaration).String), ", "))
	}
	return fmt.Sprintf("func (%s)%s", parameters, returns)
}

func (f *Function) String() string {
	var sb strings.Builder
	if f.Doc != nil {
		sb.WriteString("// ")
		sb.WriteString(*f.Doc)
		sb.WriteString("\n")
	}

	sb.WriteString("func ")
	sb.WriteString(f.Declaration())
	sb.WriteString(" {\n")

	sb.WriteString("\t<")
	sb.WriteString(f.Body)
	sb.WriteString(">\n}\n")
	return sb.String()
}

func (r *FunctionReceiver) String() string {
	var sb strings.Builder
	if r.Name != nil {
		sb.WriteString(*r.Name)
		sb.WriteString(" ")
	}
	if r.isRef {
		sb.WriteString("*")
	}
	sb.WriteString(r.RecvType)
	return sb.String()
}
