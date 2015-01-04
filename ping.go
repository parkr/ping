package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/zenazn/goji"
)

func ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "")

	referrer := r.Referer()
	if referrer == "" {
		return
	}

	url, err := url.Parse(referrer)

	if err != nil {
		http.Error(w, "Couldn't parse "+referrer+": "+err.Error(), 500)
		return
	}

	var ip string
	if os.Getenv("PING_PROXIED") == "true" {
		ip = r.Header.Get("X-Forwarded-For")
	} else {
		ip = r.RemoteAddr
	}

	visit := &Visit{
		IP:        ip,
		Host:      url.Host,
		Path:      url.Path,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	log.Println("Logging visit:", visit.String())

	err = visit.Save()

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func main() {
	goji.Get("/ping", ping)
	goji.Get("/ping.js", ping)
	goji.Serve()
}
