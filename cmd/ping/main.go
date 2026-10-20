package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/parkr/ping"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	var binding string
	flag.StringVar(&binding, "http", ":"+port, "The IP/port to bind to.")
	var hostAllowlist string
	flag.StringVar(&hostAllowlist, "hosts", "", "The hosts allowed to use this service. Comma-separated.")
	var pingBaseURL string
	flag.StringVar(&pingBaseURL, "baseurl", "http://localhost:"+port, "Base URL used for XHR request in stats.js")
	flag.Parse()

	// Opens the database, creating the schema if it is missing (idempotent).
	// Fail loudly if that can't be done: otherwise the server listens with no
	// database and every request errors, and a service manager sees it as up.
	if err := ping.Initialize(os.Getenv("PING_DB")); err != nil {
		log.Fatalf("error opening database %q: %v", os.Getenv("PING_DB"), err)
	}

	allowedHosts := strings.Split(hostAllowlist, ",")
	log.Printf("allowing the following hosts: %v", allowedHosts)

	log.Printf("base url: %q", pingBaseURL)

	http.Handle("/", ping.NewHandler(allowedHosts, pingBaseURL))

	log.Println("Listening on", binding, "...")
	log.Fatal(http.ListenAndServe(binding, nil))
}
