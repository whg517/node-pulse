package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Prompt for password
	var password string
	fmt.Print("Enter password: ")
	if _, err := fmt.Scanln(&password); err != nil {
		log.Fatalf("Failed to read password: %v", err)
	}

	// Generate bcrypt hash (cost factor 12, same as the API)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Fatalf("Failed to generate hash: %v", err)
	}

	// Output the hash
	fmt.Printf("\nPassword Hash (copy this):\n%s\n", string(hash))
}
