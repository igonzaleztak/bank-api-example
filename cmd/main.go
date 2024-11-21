package main

import (
	"bank_test/cmd/bootstrap"
	"log"
)

func main() {
	if err := bootstrap.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
