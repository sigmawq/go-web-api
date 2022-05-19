package main

import (
	"fmt"
	// "github.com/google/uuid"
)

func main() {
	fmt.Println("*** API Server *** (enjoy)")
	fmt.Println("Click enter to exit.")

	initializeStorage()
	initializeServer()
}
