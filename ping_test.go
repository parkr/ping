package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/parkr/ping/analytics"
	"github.com/parkr/ping/database"
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

	if db == nil {
		db = database.Initialize()
	}

	visitCountStart, _ := analytics.ViewsForPath(db, "/root")

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

	visitCountEnd, _ := analytics.ViewsForPath(db, "/root")

	if visitCountEnd <= visitCountStart {
		t.Errorf("visit was not saved, got %v want %v",
			visitCountEnd, visitCountStart+1)
	}
}

func TestCountsMissingParam(t *testing.T) {
	request, err := http.NewRequest("POST", "/counts", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(counts)
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	expected := "Missing param\n"

	if recorder.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got '%v' want %v",
			recorder.Body.String(), expected)
	}
}

func TestCountsValid(t *testing.T) {
	request, err := http.NewRequest("POST", "/counts", strings.NewReader("path=/root"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(counts)
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var body map[string]int
	json.NewDecoder(recorder.Body).Decode(&body)

	if val, ok := body["views"]; ok {
		if ok {
			if val < 1 {
				t.Errorf("Counts should have returned at least 1 view")
			}
		} else {
			t.Errorf("Counts did not return views")
		}
	}

	if val, ok := body["visitors"]; ok {
		if ok {
			if val < 1 {
				t.Errorf("Counts should have returned at least 1 visitor")
			}
		} else {
			t.Errorf("Counts did not return visitors")
		}
	}

}

func TestAllHost(t *testing.T) {
	request, err := http.NewRequest("POST", "/all", strings.NewReader("type=host"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(all)
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var body map[string][]string
	json.NewDecoder(recorder.Body).Decode(&body)

	expected := "example.org"
	firstElement := body["entries"][0]

	if firstElement != expected {
		t.Errorf("handler returned unexpected body: got '%v' want %v",
			firstElement, expected)
	}
}

func TestAllPath(t *testing.T) {
	request, err := http.NewRequest("POST", "/all", strings.NewReader("type=path"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(all)
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var body map[string][]string
	json.NewDecoder(recorder.Body).Decode(&body)

	expected := "/root"
	firstElement := body["entries"][0]

	if firstElement != expected {
		t.Errorf("handler returned unexpected body: got '%v' want %v",
			firstElement, expected)
	}
}
