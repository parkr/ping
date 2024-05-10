package ping

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/parkr/ping/analytics"
	"github.com/parkr/ping/database"
	"github.com/parkr/ping/dnt"
	"github.com/parkr/ping/secgpc"
)

func TestPingV2_EmptyReferrer(t *testing.T) {
	request, err := http.NewRequest("GET", "/ping?v=2", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{"example.org"}, "")
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

func TestPingV2_UnauthorizedHost(t *testing.T) {
	request, err := http.NewRequest("GET", "/ping?v=2", nil)
	if err != nil {
		t.Fatal(err)
	}

	request.Header.Set("Referer", "http://mehehe.org")

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{"example.org"}, "")
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

func TestPingV2_RequestNotToTrack(t *testing.T) {
	request, err := http.NewRequest("GET", "/ping?v=2", nil)
	if err != nil {
		t.Fatal(err)
	}

	request.Header.Set("Referer", "http://example.org")
	request.Header.Set("User-Agent", "go test client")
	request.Header.Set(dnt.DoNotTrackHeaderName, dnt.DoNotTrackHeaderValue)

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNoContent)
	}

	actual := recorder.Header().Get(dnt.DoNotTrackHeaderName)
	if actual != dnt.DoNotTrackHeaderValue {
		t.Errorf("Expected %s: %s, got: %v", dnt.DoNotTrackHeaderName, dnt.DoNotTrackHeaderValue, actual)
	}
}

func TestPingV2_RequestGlobalSecurityControl(t *testing.T) {
	request, err := http.NewRequest("GET", "/ping?v=2", nil)
	if err != nil {
		t.Fatal(err)
	}

	request.Header.Set("Referer", "http://example.org")
	request.Header.Set("User-Agent", "go test client")
	request.Header.Set(secgpc.SecGPCHeaderName, secgpc.SecGPCHeaderValue)

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNoContent)
	}

	actual := recorder.Header().Get(secgpc.SecGPCHeaderName)
	if actual != secgpc.SecGPCHeaderValue {
		t.Errorf("Expected %s: %s, got: %v", secgpc.SecGPCHeaderName, secgpc.SecGPCHeaderValue, actual)
	}
}

// BUG(jussi): Might crash because database check in database.go's init() will most
// likely return true because `checkIfSchemaExists` query doesn't include DB
func TestPingV2_Success(t *testing.T) {
	var err error
	db, err = database.InitializeForTest()
	if err != nil {
		t.Fatalf("unexpected error initializing database: %+v", err)
	}

	visitCountStart, _ := analytics.ViewsForHostPath(db, "example.org", "/root")

	request, err := http.NewRequest("GET", "/ping?v=2", nil)
	if err != nil {
		t.Fatal(err)
	}

	request.Header.Set("Referer", "http://example.org/root")
	request.Header.Set("User-Agent", "go test client")

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	assertStatusCode(t, recorder, http.StatusOK)

	expected := "function logVisit"
	if !strings.Contains(recorder.Body.String(), expected) {
		t.Errorf("pingv2 body does not contain expected string %q, got: %v",
			expected, recorder.Body.String())
	}

	visitCountEnd, _ := analytics.ViewsForHostPath(db, "example.org", "/root")

	if visitCountEnd != visitCountStart {
		t.Errorf("visit was saved but shouldn't have been, got %v want %v",
			visitCountEnd, visitCountStart)
	}
}

func TestSubmitV2_MissingHost(t *testing.T) {
	request, err := http.NewRequest("POST", "/submit.js", strings.NewReader("host=&path=/root"))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "go test client")
	request.Header.Set("Referer", "https://example.org")

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	expected := "(function(){console.error(\"missing param\")})();"
	if recorder.Body.String() != expected {
		t.Errorf("submitv2 body is not expected string %q, got: %v",
			expected, recorder.Body.String())
	}

	verifyCorsHeaders(t, recorder, "https://example.org")
}

func TestSubmitV2_MissingPath(t *testing.T) {
	request, err := http.NewRequest("POST", "/submit.js", strings.NewReader("host=example.org&path="))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "go test client")
	request.Header.Set("Referer", "https://example.org")

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	expected := "(function(){console.error(\"missing param\")})();"
	if recorder.Body.String() != expected {
		t.Errorf("submitv2 body is not expected string %q, got: %v",
			expected, recorder.Body.String())
	}

	verifyCorsHeaders(t, recorder, "https://example.org")
}

func TestSubmitV2_InvalidHost(t *testing.T) {
	request, err := http.NewRequest("POST", "/submit.js", strings.NewReader(`host=\\.\\%21...boom%21%21&path=/root`))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "go test client")
	request.Header.Set("Referer", "https://example.org")

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}

	expected := "(function(){console.error(\"Couldn't parse referrer: parse \"//%!C(MISSING)%!C(MISSING).%!C(MISSING)%!C(MISSING)!...boom!!/root\": invalid URL escape \"%!C(MISSING)\"\")})();"
	if recorder.Body.String() != expected {
		t.Errorf("submitv2 body is not expected string %q, got: %v",
			expected, recorder.Body.String())
	}
}

func TestSubmitV2_UnauthorizedHost(t *testing.T) {
	request, err := http.NewRequest("POST", "/submit.js", strings.NewReader("host=unauthorized.org&path=/root"))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "go test client")
	request.Header.Set("Referer", "https://example.org")

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
	}

	expected := "(function(){console.error(\"unauthorized host\")})();"
	if recorder.Body.String() != expected {
		t.Errorf("submitv2 body is not expected string %q, got: %v",
			expected, recorder.Body.String())
	}

	verifyCorsHeaders(t, recorder, "https://example.org")
}

func TestSubmitV2_MissingUserAgent(t *testing.T) {
	request, err := http.NewRequest("POST", "/submit.js", strings.NewReader("host=example.org&path=/root"))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "") // missing
	request.Header.Set("Referer", "https://example.org")

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	expected := "(function(){console.error(\"empty user-agent\")})();"
	if recorder.Body.String() != expected {
		t.Errorf("submitv2 body is not expected string %q, got: %v",
			expected, recorder.Body.String())
	}

	verifyCorsHeaders(t, recorder, "https://example.org")
}

func TestSubmitV2_Success(t *testing.T) {
	var err error
	db, err = database.InitializeForTest()
	if err != nil {
		t.Fatalf("unexpected error initializing database: %+v", err)
	}

	visitCountStart, _ := analytics.ViewsForHostPath(db, "example.org", "/TestSubmitV2_Success")

	request, err := http.NewRequest("POST", "/submit.js", strings.NewReader("host=example.org&path=/TestSubmitV2_Success"))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "go test client")
	request.Header.Set("Referer", "https://example.org/")
	request.RemoteAddr = "100.0.0.0"

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	if status := recorder.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	expected := "(function(){})();"
	if recorder.Body.String() != expected {
		t.Errorf("submitv2 body is not expected string %q, got: %v",
			expected, recorder.Body.String())
	}

	verifyCorsHeaders(t, recorder, "https://example.org")

	visitCountEnd, _ := analytics.ViewsForHostPath(db, "example.org", "/TestSubmitV2_Success")

	if visitCountEnd != visitCountStart+1 {
		t.Errorf("visit wasn't saved, got %v want %v",
			visitCountEnd, visitCountStart+1)
	}

	visit, err := database.Get(db, 1)
	if err != nil {
		t.Errorf("expected no error getting visit from db, got: %v", err)
	}

	expectedUserAgent := "go test client"
	if visit.UserAgent != expectedUserAgent {
		t.Errorf("expected visit user agent %q, got %v", expectedUserAgent, visit.UserAgent)
	}

	expectedIP := "100.0.0.0"
	if visit.IP != expectedIP {
		t.Errorf("expected visit ip %q, got: %v", expectedIP, visit.IP)
	}
}

func TestSubmitV2_Success_PreservesXForwardedForOverRemoteAddr(t *testing.T) {
	var err error
	db, err = database.InitializeForTest()
	if err != nil {
		t.Fatalf("unexpected error initializing database: %+v", err)
	}

	request, err := http.NewRequest("POST", "/submit.js", strings.NewReader("host=example.org&path=/TestSubmitV2_Success"))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "go test client")
	request.Header.Set("Referer", "https://example.org/")
	request.Header.Set(xForwardedForHeaderName, "100.0.12.0")
	request.RemoteAddr = "127.0.0.1:12324"

	recorder := httptest.NewRecorder()
	handler := NewHandler([]string{"example.org"}, "")
	handler.ServeHTTP(recorder, request)

	assertStatusCode(t, recorder, http.StatusCreated)

	expected := "(function(){})();"
	if recorder.Body.String() != expected {
		t.Errorf("submitv2 body is not expected string %q, got: %v",
			expected, recorder.Body.String())
	}

	verifyCorsHeaders(t, recorder, "https://example.org")

	visit, err := database.Get(db, 1)
	if err != nil {
		t.Errorf("expected no error getting visit from db, got: %v", err)
	}

	expectedIP := "100.0.12.0"
	if visit.IP != expectedIP {
		t.Errorf("expected visit ip %q, got: %v", expectedIP, visit.IP)
	}
}
