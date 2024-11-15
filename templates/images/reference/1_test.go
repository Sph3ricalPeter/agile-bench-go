package main

import "testing"

func TestModels(t *testing.T) {
	address := Address{
		Line1:    "123 Fake St",
		Line2:    "Apt 1",
		Postcode: "12345",
	}
	_ = Person{
		Name:    "John Doe",
		Age:     30,
		Address: address,
	}
}
