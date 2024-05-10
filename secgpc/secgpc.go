package secgpc

import (
	"net/http"
)

const SecGPCHeaderName = "Sec-GPC"
const SecGPCHeaderValue = "1"

func RequestsGlobalPrivacyControl(r *http.Request) bool {
	return r.Header.Get(SecGPCHeaderName) == SecGPCHeaderValue
}

func SetGlobalPrivacyControl(w http.ResponseWriter) {
	w.Header().Set(SecGPCHeaderName, SecGPCHeaderValue)
}

func NewMiddleware(nextHandler http.Handler) http.Handler {
	return secgpcMiddleware{nextHandler: nextHandler}
}

type secgpcMiddleware struct {
	nextHandler http.Handler
}

func (d secgpcMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if RequestsGlobalPrivacyControl(r) {
		SetGlobalPrivacyControl(w)
		// Note: All w.Header() modifications must be made BEFORE this call.
		w.WriteHeader(http.StatusNoContent)
		return
	}
	d.nextHandler.ServeHTTP(w, r)
}
