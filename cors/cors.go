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
		w.WriteHeader(http.StatusNoContent)
		c.addCORSHeaders(w, r)
		return
	}
	c.next.ServeHTTP(w, r)
	c.addCORSHeaders(w, r)
}

func (c corsHandler) addCORSHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(CorsAccessControlAllowMethodsHeaderName, "GET, POST")
	if sanitizedOrigin, ok := c.allowCORSOrigin(r.Header.Get("Origin")); ok {
		w.Header().Set(CorsAccessControlAllowOriginHeaderName, sanitizedOrigin)
	} else if sanitizedOrigin, ok := c.allowCORSOrigin(r.Referer()); ok {
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
	return parsedOrigin.String(), c.allowedHost(originHostname)
}

func (c corsHandler) allowedHost(hostname string) bool {
	_, ok := c.allowedHosts[hostname]
	return ok
}
