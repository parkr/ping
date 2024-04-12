package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/parkr/ping/analytics"
	"github.com/parkr/ping/database"
	"github.com/parkr/ping/dnt"
)

// Tests white listing.
// Hosts can be allowlisted so other sites can't mess with page views.
func TestAllowedHost(t *testing.T) {
	*hostAllowlist = "example.org"

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
	handler := buildHandler()
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
	handler := buildHandler()
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusForbidden)
	}

	expected := `(function(){console.error("unauthorized host")})();`

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
	handler := buildHandler()
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

func TestPingRequestNotToTrack(t *testing.T) {
	request, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	request.Header.Set("Referer", "http://example.org")
	request.Header.Set("User-Agent", "go test client")
	request.Header.Set(dnt.DoNotTrackHeaderName, dnt.DoNotTrackHeaderValue)

	recorder := httptest.NewRecorder()
	handler := buildHandler()
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	expected := `(function(){})();`

	if recorder.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			recorder.Body.String(), expected)
	}

	actual := recorder.Header().Get(dnt.DoNotTrackHeaderName)
	if actual != dnt.DoNotTrackHeaderValue {
		t.Errorf("Expected %s: %s, got: %v", dnt.DoNotTrackHeaderName, dnt.DoNotTrackHeaderValue, actual)
	}
}

// BUG(jussi): Might crash because database check in database.go's init() will most
// likely return true because `checkIfSchemaExists` query doesn't include DB
func TestPingSuccess(t *testing.T) {
	*hostAllowlist = "example.org"

	if db == nil {
		var err error
		db, err = database.Initialize()
		if err != nil {
			t.Fatalf("unexpected error initializing database: %+v", err)
		}
	}

	visitCountStart, _ := analytics.ViewsForHostPath(db, "example.org", "/root")

	request, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	request.Header.Set("Referer", "http://example.org/root")
	request.Header.Set("User-Agent", "go test client")

	recorder := httptest.NewRecorder()
	handler := buildHandler()
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

	visitCountEnd, _ := analytics.ViewsForHostPath(db, "example.org", "/root")

	if visitCountEnd <= visitCountStart {
		t.Errorf("visit was not saved, got %v want %v",
			visitCountEnd, visitCountStart+1)
	}
}

func TestCountsOptionsPreflight(t *testing.T) {
	*hostAllowlist = "example.org"

	request, err := http.NewRequest(http.MethodOptions, "/counts", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Add("Origin", "https://example.org")

	recorder := httptest.NewRecorder()
	handler := buildHandler()
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNoContent)
	}

	expectedAllowedHosts := "https://example.org"
	actual := recorder.Header().Get("Access-Control-Allow-Origin")
	if actual != expectedAllowedHosts {
		t.Errorf("expected Access-Control-Allow-Origin: %v, got: %v", expectedAllowedHosts, actual)
	}

	expectedAllowedMethods := "GET"
	actual = recorder.Header().Get("Access-Control-Allow-Methods")
	if actual != expectedAllowedMethods {
		t.Errorf("expected Access-Control-Allow-Methods: %v, got: %v", expectedAllowedMethods, actual)
	}
}

func TestCountsMissingParam(t *testing.T) {
	request, err := http.NewRequest("POST", "/counts", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := buildHandler()
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

func TestCountsMissingHostParam(t *testing.T) {
	request, err := http.NewRequest("POST", "/counts", strings.NewReader("path=/root"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := buildHandler()
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
	request, err := http.NewRequest("POST", "/counts", strings.NewReader("path=/root&host=example.org"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := buildHandler()
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

func TestAllOptionsPreflight(t *testing.T) {
	request, err := http.NewRequest(http.MethodOptions, "/all", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Add("Origin", "https://example.org")

	recorder := httptest.NewRecorder()
	handler := buildHandler()
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNoContent)
	}

	expectedAllowedHosts := "https://example.org"
	actual := recorder.Header().Get("Access-Control-Allow-Origin")
	if actual != expectedAllowedHosts {
		t.Errorf("expected Access-Control-Allow-Origin: %v, got: %v", expectedAllowedHosts, actual)
	}

	expectedAllowedMethods := "GET"
	actual = recorder.Header().Get("Access-Control-Allow-Methods")
	if actual != expectedAllowedMethods {
		t.Errorf("expected Access-Control-Allow-Methods: %v, got: %v", expectedAllowedMethods, actual)
	}
}

func TestAllHost(t *testing.T) {
	request, err := http.NewRequest("POST", "/all", strings.NewReader("type=host"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := buildHandler()
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var body map[string][]string
	json.NewDecoder(recorder.Body).Decode(&body)

	if len(body["entries"]) < 1 {
		t.Errorf("expected 'entries' to exist, but was '%#v'", body["entries"])
		return
	}

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
	handler := buildHandler()
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var body map[string][]string
	json.NewDecoder(recorder.Body).Decode(&body)

	if len(body["entries"]) < 1 {
		t.Errorf("expected 'entries' to exist, but was '%#v'", body["entries"])
		return
	}

	expected := "/root"
	firstElement := body["entries"][0]

	if firstElement != expected {
		t.Errorf("handler returned unexpected body: got '%v' want %v",
			firstElement, expected)
	}
}
