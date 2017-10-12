package main

import (
	"net/http"
	"net/textproto"
	"testing"
)

func TestRequestsDoNotTrack(t *testing.T) {
	// Header is set correctly. Do not track this request.
	req := &http.Request{Header: http.Header(map[string][]string{
		// We use Dnt here instead of DNT because of the way req.Header.Get
		// works. It converts the input with CanonicalMIMEHeaderKey and
		// checks the map for that value. When the request is built, the
		// map is made by converting keys using this same method so the
		// lookup is always seamless.
		"Dnt": {DoNotTrackHeaderValue},
	})}

	if !requestsDoNotTrack(req) {
		t.Fatalf("expected DNT header to be respected, headers: %+v, %s header: %+v",
			req.Header,
			textproto.CanonicalMIMEHeaderKey(DoNotTrackHeaderName),
			req.Header.Get(DoNotTrackHeaderName))
	}

	// Header is not set. We can track this request.
	req = &http.Request{Header: map[string][]string{}}

	if requestsDoNotTrack(req) {
		t.Fatalf("expected lack of DNT header to mean we can track")
	}

	// Header is not set correctly. We can track this request.
	req = &http.Request{Header: map[string][]string{
		"Dnt": []string{"!"},
	}}

	if requestsDoNotTrack(req) {
		t.Fatalf("expected value of DNT header to mean we can track")
	}
}
