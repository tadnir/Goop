package APackage

//go:generate go run github.com/tadnir/goop
type A struct {
	aVtable   `goop:"vtable"`
	firstName string
}

func (a *A) New(firstName string) *A {
	a.initClass()
	a.firstName = firstName
	return a
}

func (a *A) getNameImpl() string {
	return a.firstName
}

func (a *A) Foo() {
	// …

	// this calls the virtual function held in the vtable at the moment of invocation
	println(a.getName())

	// …
}
