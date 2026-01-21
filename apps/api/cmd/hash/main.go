package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run hash.go <password>")
		os.Exit(1)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(os.Args[1]), 10)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Println(string(hash))
}
