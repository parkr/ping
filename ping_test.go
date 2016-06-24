package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Tests white listing.
// Hosts can be whitelisted so other sites can't mess with page views.
func TestAllowedHost(t *testing.T) {
	*whitelist = "example.org"

	if allowedHost("badexample.org") {
		t.Error("Host badexample.org shouldn't be allowed to access")
	}
}

func TestPingEmptyReferrer(t *testing.T) {
	request, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(ping)
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	expected := `(function(){console.error("empty referrer")})();`

	if recorder.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			recorder.Body.String(), expected)
	}

}

func TestPingUnauthorizedHost(t *testing.T) {
	request, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	request.Header.Set("Referer", "http://mehehe.org")

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(ping)
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusForbidden {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusForbidden)
	}

	expected := `(function(){console.error("love the host, except noooope.")})();`

	if recorder.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			recorder.Body.String(), expected)
	}
}

func TestPingEmptyUserAgent(t *testing.T) {
	request, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	request.Header.Set("Referer", "http://example.org")

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(ping)
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	expected := `(function(){console.error("empty user-agent")})();`

	if recorder.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			recorder.Body.String(), expected)
	}
}

// BUG(jussi): Might crash because database check in database.go's init() will most
//             likely return true because `checkIfSchemaExists` query doesn't include DB
func TestPingSuccess(t *testing.T) {
	*whitelist = "example.org"

	request, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	request.Header.Set("Referer", "http://example.org/root")
	request.Header.Set("User-Agent", "go test client")

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(ping)
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	expected := `(function(){})();`

	if recorder.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			recorder.Body.String(), expected)
	}
}
