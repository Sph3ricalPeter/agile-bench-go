package main

type Person struct {
	Name    string  `json:"name"`
	Age     int     `json:"age"`
	Address Address `json:"address"`
}

type Address struct {
	Line1    string `json:"line1"`
	Line2    string `json:"line2"`
	Postcode string `json:"postcode"`
}

func main() {
}
