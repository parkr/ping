package ping

import (
	"log"
	"net/http"
	"net/url"

	"github.com/parkr/ping/jsv1"
)

func NewHostAuthMiddleware(allowedHosts []string, nextHandler http.Handler) http.Handler {
	allowedHostsMap := make(map[string]bool, len(allowedHosts))
	for _, allowedHost := range allowedHosts {
		allowedHostsMap[allowedHost] = true
	}

	return hostAuthMiddleware{
		allowedHosts: allowedHostsMap,
		next:         nextHandler,
	}
}

type hostAuthMiddleware struct {
	allowedHosts map[string]bool
	next         http.Handler
}

func (m hostAuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	referrer := r.Referer()
	if referrer == "" {
		log.Println("empty referrer")
		jsv1.Error(w, http.StatusBadRequest, "empty referrer")
		return
	}

	url, err := url.Parse(referrer)

	if err != nil {
		log.Println("invalid referrer:", sanitizeUserInput(referrer))
		jsv1.Error(w, http.StatusInternalServerError, "Couldn't parse referrer: "+err.Error())
		return
	}

	if !m.allowedHost(url.Host) {
		log.Println("unauthorized host:", sanitizeUserInput(url.Host))
		jsv1.Error(w, http.StatusUnauthorized, "unauthorized host")
		return
	}

	m.next.ServeHTTP(w, r)
}

func (m hostAuthMiddleware) allowedHost(hostname string) bool {
	_, ok := m.allowedHosts[hostname]
	return ok
}
