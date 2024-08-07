package ping

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/parkr/ping/analytics"
	"github.com/parkr/ping/cors"
	"github.com/parkr/ping/database"
	"github.com/parkr/ping/dnt"
	"github.com/parkr/ping/secgpc"
)

func TestPingEmptyReferrer(t *testing.T) {
	request, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	assertStatusCode(t, recorder, http.StatusBadRequest)

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
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	assertStatusCode(t, recorder, http.StatusUnauthorized)

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
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	assertStatusCode(t, recorder, http.StatusBadRequest)

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
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	assertStatusCode(t, recorder, http.StatusNoContent)

	actual := recorder.Header().Get(dnt.DoNotTrackHeaderName)
	if actual != dnt.DoNotTrackHeaderValue {
		t.Errorf("Expected %s: %s, got: %v", dnt.DoNotTrackHeaderName, dnt.DoNotTrackHeaderValue, actual)
	}
}

func TestPingRequestGlobalPrivacyControl(t *testing.T) {
	request, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	request.Header.Set("Referer", "http://example.org")
	request.Header.Set("User-Agent", "go test client")
	request.Header.Set(secgpc.SecGPCHeaderName, secgpc.SecGPCHeaderValue)

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	assertStatusCode(t, recorder, http.StatusNoContent)

	actual := recorder.Header().Get(secgpc.SecGPCHeaderName)
	if actual != secgpc.SecGPCHeaderValue {
		t.Errorf("Expected %s: %s, got: %v", secgpc.SecGPCHeaderName, secgpc.SecGPCHeaderValue, actual)
	}
}

// BUG(jussi): Might crash because database check in database.go's init() will most
// likely return true because `checkIfSchemaExists` query doesn't include DB
func TestPingSuccess(t *testing.T) {
	var err error
	db, err = database.InitializeForTest()
	if err != nil {
		t.Fatalf("unexpected error initializing database: %+v", err)
	}

	visitCountStart, _ := analytics.ViewsForHostPath(db, "example.org", "/root")

	request, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	request.Header.Set("Referer", "http://example.org/root")
	request.Header.Set("User-Agent", "go test client")

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	assertStatusCode(t, recorder, http.StatusCreated)

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
	request, err := http.NewRequest(http.MethodOptions, "/counts", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Add("Origin", "https://example.org")

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	assertStatusCode(t, recorder, http.StatusNoContent)
	verifyCorsHeaders(t, recorder, "https://example.org")
}

func TestCountsMissingParam(t *testing.T) {
	request, err := http.NewRequest("POST", "/counts", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	assertStatusCode(t, recorder, http.StatusBadRequest)

	expected := `{"error":"missing param"}` + "\n"

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
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	assertStatusCode(t, recorder, http.StatusBadRequest)

	expected := `{"error":"missing param"}` + "\n"

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
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	assertStatusCode(t, recorder, http.StatusOK)

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
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	assertStatusCode(t, recorder, http.StatusNoContent)

	verifyCorsHeaders(t, recorder, "https://example.org")
}

func TestAllHost(t *testing.T) {
	request, err := http.NewRequest("POST", "/all", strings.NewReader("type=host"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	assertStatusCode(t, recorder, http.StatusOK)

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
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	assertStatusCode(t, recorder, http.StatusOK)

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

func TestStats_Success(t *testing.T) {
	pingBaseURL := "http://ping.mywebsite.com"

	request, err := http.NewRequest("GET", "/stats.js", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Referer", pingBaseURL+"/foobar")

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{"ping.mywebsite.com"}, pingBaseURL)
	handler.ServeHTTP(recorder, request)

	assertStatusCode(t, recorder, http.StatusOK)
	verifyCorsHeaders(t, recorder, pingBaseURL)

	allExpectedBodyContents := []string{
		"fetchPingStats",
		"document.addEventListener",
		"writePingStatsToHTML",
		"http://ping.mywebsite.com/counts",
	}
	for _, expectedBodyContents := range allExpectedBodyContents {
		if !strings.Contains(recorder.Body.String(), expectedBodyContents) {
			t.Errorf("handler returned body which does not contain %q: %s",
				expectedBodyContents, recorder.Body.String())
		}
	}
}

func TestHealth_NoDB(t *testing.T) {
	db = nil

	request, err := http.NewRequest("GET", "/_health", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{}, "")
	handler.ServeHTTP(recorder, request)

	assertStatusCode(t, recorder, http.StatusInternalServerError)
}

func TestHealth_DB(t *testing.T) {
	var err error
	db, err = database.InitializeForTest()
	if err != nil {
		t.Errorf("unable to initialize database: %v", err)
	}

	request, err := http.NewRequest("GET", "/_health", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{}, "")
	handler.ServeHTTP(recorder, request)

	assertStatusCode(t, recorder, http.StatusOK)
}

func assertStatusCode(t *testing.T, recorder *httptest.ResponseRecorder, expectedCode int) {
	if recorder.Code != expectedCode {
		t.Errorf("handler expected status code %d, got %v",
			expectedCode, recorder.Code)
	}
}

func verifyCorsHeaders(t *testing.T, recorder *httptest.ResponseRecorder, origin string) {
	actual := recorder.Header().Get(cors.CorsAccessControlAllowOriginHeaderName)
	if actual != origin {
		t.Errorf("expected %s: %v, got: %v", cors.CorsAccessControlAllowOriginHeaderName, origin, actual)
	}

	expectedAllowedMethods := "GET, POST"
	actual = recorder.Header().Get(cors.CorsAccessControlAllowMethodsHeaderName)
	if actual != expectedAllowedMethods {
		t.Errorf("expected %s: %v, got: %v", cors.CorsAccessControlAllowMethodsHeaderName, expectedAllowedMethods, actual)
	}
}
