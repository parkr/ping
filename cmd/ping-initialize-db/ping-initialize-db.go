package main

import (
	"log"
	"os"

	"github.com/parkr/ping/database"
)

func VerifySchema() error {
	_, err := database.Initialize()
	return err
}

func main() {
	if err := VerifySchema(); err != nil {
		log.Fatalf("error setting up database: %+v", err)
	}
	log.Printf("database setup at %s", os.Getenv("PING_DB"))
}
