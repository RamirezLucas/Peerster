package main

import (
	"fmt"
	"os"
)

func main() {

	// Open the file
	f, err := os.Open("../_SharedFiles/tmp.txt")
	if err != nil {
		fmt.Printf("aaa")
		return
	}

	b1 := make([]byte, 7)
	n1, err := f.Read(b1)
	fmt.Printf("%d bytes: %s\n", n1, string(b1))

	b2 := make([]byte, 7)
	n2, err := f.Read(b2)
	fmt.Printf("%d bytes: %s\n", n2, string(b2))

}
