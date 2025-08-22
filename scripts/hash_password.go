package main

import (
	"fmt"
	"log"
	"os"
	
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: go run hash_password.go <password>")
	}
	
	password := os.Args[1]
	
	// Use cost 10 which is the same as bcrypt.DefaultCost
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Error hashing password: %v", err)
	}
	
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Hash: %s\n", string(hashedPassword))
}