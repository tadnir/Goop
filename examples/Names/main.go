package main

import "goscratch/APackage"

func main() {
	a := new(APackage.A).New("John")
	a.Foo() // prints "John"

	b := new(APackage.B).New("John", "Doe")
	b.Foo() // prints "John Doe"

	c := new(APackage.C).New("John", "Jimmy", "Doe")
	c.Foo() // prints "John Doe"
}
