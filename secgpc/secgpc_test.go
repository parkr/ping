package secgpc

import (
	"net/http"
	"net/textproto"
	"testing"
)

func TestRequestsGlobalPrivacyControl(t *testing.T) {
	// Header is set correctly. Do not track this request.
	req := &http.Request{Header: http.Header(map[string][]string{
		// We use Sec-Gpc here instead of Sec-GPC because of the way req.Header.Get
		// works. It converts the input with CanonicalMIMEHeaderKey and
		// checks the map for that value. When the request is built, the
		// map is made by converting keys using this same method so the
		// lookup is always seamless.
		"Sec-Gpc": {SecGPCHeaderValue},
	})}

	if !RequestsGlobalPrivacyControl(req) {
		t.Fatalf("expected Sec-GPC header to be respected, headers: %+v, %s header: %+v",
			req.Header,
			textproto.CanonicalMIMEHeaderKey(SecGPCHeaderName),
			req.Header.Get(SecGPCHeaderName))
	}

	// Header is not set. We can track this request.
	req = &http.Request{Header: map[string][]string{}}

	if RequestsGlobalPrivacyControl(req) {
		t.Fatalf("expected lack of Sec-GPC header to mean we can track")
	}

	// Header is not set correctly. We can track this request.
	req = &http.Request{Header: map[string][]string{
		"Sec-Gpc": {"!"},
	}}

	if RequestsGlobalPrivacyControl(req) {
		t.Fatalf("expected value of Sec-GPC header to mean we can track")
	}
}
