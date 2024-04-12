package dnt

import "net/http"

const DoNotTrackHeaderName = "DNT"
const DoNotTrackHeaderValue = "1"

func RequestsDoNotTrack(r *http.Request) bool {
	return r.Header.Get(DoNotTrackHeaderName) == DoNotTrackHeaderValue
}

func SetDoNotTrack(w http.ResponseWriter) {
	w.Header().Set(DoNotTrackHeaderName, DoNotTrackHeaderValue)
}
