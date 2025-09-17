//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// 預設測試密碼: Test123456!
// 用法: go run verify_password.go <hash> <password>
func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: go run verify_password.go <hash> <password>")
	}

	hash := os.Args[1]
	password := os.Args[2]

	// 比較雜湊值和密碼
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		fmt.Printf("密碼驗證失敗: %v\n", err)
		fmt.Printf("Hash: %s\n", hash)
		fmt.Printf("Password: %s\n", password)
		os.Exit(1)
	}

	fmt.Printf("✅ 密碼驗證成功!\n")
	fmt.Printf("Hash: %s\n", hash)
	fmt.Printf("Password: %s\n", password)
}
