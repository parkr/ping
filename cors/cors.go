package cors

import (
	"log"
	"net/http"
	"net/url"
)

const (
	CorsAccessControlAllowMethodsHeaderName = "Access-Control-Allow-Methods"
	CorsAccessControlAllowOriginHeaderName  = "Access-Control-Allow-Origin"
)

func NewMiddleware(allowedHosts []string, nextHandler http.Handler) corsHandler {
	allowedHostsMap := make(map[string]bool, len(allowedHosts))
	for _, allowedHost := range allowedHosts {
		allowedHostsMap[allowedHost] = true
	}

	return corsHandler{
		allowedHosts: allowedHostsMap,
		next:         nextHandler,
	}
}

type corsHandler struct {
	allowedHosts map[string]bool
	next         http.Handler
}

// ServeHTTP adds CORS headers.
func (c corsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		// According to the documentation for ResponseWriter.WriteHeader and
		// ResponseWriter.Header, modifications to the Header must happen BEFORE
		// any call to WriteHeader. That means it should be, in order:
		// 1. Write all your Headers
		// 2. Call WriteHeader to set your status
		// 3. Write to the body
		c.addCORSHeaders(w, r)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	c.addCORSHeaders(w, r)
	c.next.ServeHTTP(w, r)
}

func (c corsHandler) addCORSHeaders(w http.ResponseWriter, r *http.Request) {
	log.Printf("cors: adding headers for %q", r.URL.Path)
	w.Header().Set(CorsAccessControlAllowMethodsHeaderName, "GET, POST")
	if sanitizedOrigin, ok := c.allowCORSOrigin(r.Header.Get("Origin")); ok {
		log.Printf("cors: sanitized origin %q", sanitizedOrigin)
		w.Header().Set(CorsAccessControlAllowOriginHeaderName, sanitizedOrigin)
	} else if sanitizedOrigin, ok := c.allowCORSOrigin(r.Referer()); ok {
		log.Printf("cors: sanitized referer %q", sanitizedOrigin)
		w.Header().Set(CorsAccessControlAllowOriginHeaderName, sanitizedOrigin)
	}
}

func (c corsHandler) allowCORSOrigin(origin string) (string, bool) {
	if origin == "" {
		return "", false
	}

	parsedOrigin, err := url.Parse(origin)
	if err != nil {
		log.Printf("cors: unable to parse origin %q: %v", origin, err)
		return "", false
	}
	parsedOrigin.Path = ""

	originHostname := parsedOrigin.Hostname()
	_, ok := c.allowedHosts[originHostname]
	return parsedOrigin.String(), ok
}
