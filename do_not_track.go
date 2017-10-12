package main

import "net/http"

const DoNotTrackHeaderName = "DNT"
const DoNotTrackHeaderValue = "1"

func requestsDoNotTrack(r *http.Request) bool {
	return r.Header.Get(DoNotTrackHeaderName) == DoNotTrackHeaderValue
}
