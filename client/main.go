package main

import "fmt"

func main() {

	var client Client

	if err := gossiper.parseArgumentsGossiper(); err != nil {
		fmt.Println(err)
		return
	}

}
