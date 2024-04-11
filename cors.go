package main

import (
	"log"
	"net/http"
	"net/url"
)

type corsHandler struct {
	next http.HandlerFunc
}

// ServeHTTP adds CORS headers.
func (c *corsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		addCorsHeaders(w, r)
		return
	}
	c.next.ServeHTTP(w, r)
}

func addCorsHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	origin := r.Header.Get("Origin")
	if origin != "" {
		parsedOrigin, err := url.Parse(origin)
		if err != nil {
			log.Printf("unable to parse origin: %v", origin)
		} else {
			originHostname := parsedOrigin.Hostname()
			for _, allowedHost := range allowedHosts {
				if allowedHost == originHostname {
					w.Header().Set("Access-Control-Allow-Origin", "https://"+allowedHost)
					break
				}
			}
		}
	}
}
