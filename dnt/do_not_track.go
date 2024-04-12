package dnt

import (
	"net/http"
)

const DoNotTrackHeaderName = "DNT"
const DoNotTrackHeaderValue = "1"

func RequestsDoNotTrack(r *http.Request) bool {
	return r.Header.Get(DoNotTrackHeaderName) == DoNotTrackHeaderValue
}

func SetDoNotTrack(w http.ResponseWriter) {
	w.Header().Set(DoNotTrackHeaderName, DoNotTrackHeaderValue)
}

func NewMiddleware(nextHandler http.Handler) http.Handler {
	return dntMiddleware{nextHandler: nextHandler}
}

type dntMiddleware struct {
	nextHandler http.Handler
}

func (d dntMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if RequestsDoNotTrack(r) {
		SetDoNotTrack(w)
		// Note: All w.Header() modifications must be made BEFORE this call.
		w.WriteHeader(http.StatusNoContent)
		return
	}
	d.nextHandler.ServeHTTP(w, r)
}
