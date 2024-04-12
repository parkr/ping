package main

import (
	"log"
	"os"

	"github.com/parkr/ping/database"
)

func VerifySchema(connection string) error {
	_, err := database.Initialize(connection)
	return err
}

func main() {
	connection := os.Getenv("PING_DB")
	if err := VerifySchema(connection); err != nil {
		log.Fatalf("error setting up database: %+v", err)
	}
	log.Printf("database setup at %s", connection)
}
