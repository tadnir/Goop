package go_generator

import "fmt"

type goVarDecl struct {
	name     string
	declType string
}

func (v goVarDecl) String() string {
	return fmt.Sprintf("%v %v", v.name, v.declType)
}
