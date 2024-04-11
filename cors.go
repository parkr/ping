package main

import "net/http"

type corsHandler struct {
	next http.HandlerFunc
}

// ServeHTTP adds CORS headers.
func (c *corsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		addCorsHeaders(w)
		return
	}
	c.next.ServeHTTP(w, r)
}

func addCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	for _, allowedHost := range allowedHosts {
		w.Header().Add("Access-Control-Allow-Origin", "https://"+allowedHost)
	}
}
