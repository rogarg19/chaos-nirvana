package main

import (
	"flag"
	"fmt"
)

func main() {
	fmt.Println("Starting execution")
	testType := flag.String("type", "", "Specify test to be executed")

	flag.Parse()

	if testType == nil {
		return
	}
}
