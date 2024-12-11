package APackage

//go:generate go run github.com/tadnir/goop
type B struct {
	A        `goop:"super"`
	lastName string
}

func (b *B) New(firstName string, lastName string) *B {
	b.super().New(firstName)
	b.lastName = lastName
	return b
}

func (b *B) getNameImpl() string {
	return b.firstName + " " + b.lastName
}
