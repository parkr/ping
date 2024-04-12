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

	ping.Initialize(os.Getenv("PING_DB"))

	allowedHosts := strings.Split(hostAllowlist, ",")

	http.Handle("/", ping.NewHandler(allowedHosts, pingBaseURL))

	log.Println("Listening on", binding, "...")
	log.Fatal(http.ListenAndServe(binding, nil))
}
