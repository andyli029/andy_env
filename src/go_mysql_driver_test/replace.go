package main

import (
	"fmt"
	"strings"
)


func main() {
	var old string
	new := "a"
	s := "bbaba"

	strings.Replace(s, old, new, 1)
	fmt.Println("the new is:", s)

}
