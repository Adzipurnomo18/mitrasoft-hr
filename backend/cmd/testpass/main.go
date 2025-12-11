package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: testpass <hash> <plain-password>")
		return
	}

	hash := os.Args[1]
	plain := os.Args[2]

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	if err != nil {
		fmt.Println("NOT MATCH:", err)
		return
	}

	fmt.Println("OK: password match")
}
