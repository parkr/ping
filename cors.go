package main

import (
	"log"
	"net/http"
	"net/url"
)

const (
	CorsAccessControlAllowMethodsHeaderName = "Access-Control-Allow-Methods"
	CorsAccessControlAllowOriginHeaderName  = "Access-Control-Allow-Origin"
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
	w.Header().Set(CorsAccessControlAllowMethodsHeaderName, "GET, POST")
	if allowCORSOrigin(r.Header.Get("Origin")) {
		w.Header().Set(CorsAccessControlAllowOriginHeaderName, r.Header.Get("Origin"))
	} else if allowCORSOrigin(r.Referer()) {
		w.Header().Set(CorsAccessControlAllowOriginHeaderName, r.Referer())
	}
}

func allowCORSOrigin(origin string) bool {
	if origin == "" {
		return false
	}

	parsedOrigin, err := url.Parse(origin)
	if err != nil {
		log.Printf("cors: unable to parse origin %q: %v", origin, err)
		return false
	}

	originHostname := parsedOrigin.Hostname()
	return allowedHost(originHostname)
}
