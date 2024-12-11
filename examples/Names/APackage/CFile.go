package APackage

//go:generate go run github.com/tadnir/goop
type C struct {
	B          `goop:"super"`
	middleName string
}

func (c *C) New(firstName string, middleName string, lastName string) *C {
	c.super().New(firstName, lastName)
	c.middleName = middleName
	return c
}

func (c *C) getNameImpl() string {
	return c.firstName + " " + c.middleName + " " + c.lastName
}
