package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddCorsHeaders_OriginRequestHeader_Success(t *testing.T) {
	middleware := NewMiddleware([]string{"example.org"}, nil)

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodOptions, "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Add("Origin", "https://example.org")

	middleware.ServeHTTP(recorder, request)

	expectedAllowedMethods := "GET, POST"
	actual := recorder.Header().Get(CorsAccessControlAllowMethodsHeaderName)
	if actual != expectedAllowedMethods {
		t.Errorf("expected %s: %v, got: %v", CorsAccessControlAllowMethodsHeaderName, expectedAllowedMethods, actual)
	}

	expectedAllowedHosts := "https://example.org"
	actual = recorder.Header().Get(CorsAccessControlAllowOriginHeaderName)
	if actual != expectedAllowedHosts {
		t.Errorf("expected %s: %v, got: %v", CorsAccessControlAllowOriginHeaderName, expectedAllowedHosts, actual)
	}
}

func TestAddCorsHeaders_RefererRequestHeader_Success(t *testing.T) {
	middleware := NewMiddleware([]string{"example.org"}, nil)

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodOptions, "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Add("Referer", "https://example.org")

	middleware.ServeHTTP(recorder, request)

	expectedAllowedMethods := "GET, POST"
	actual := recorder.Header().Get(CorsAccessControlAllowMethodsHeaderName)
	if actual != expectedAllowedMethods {
		t.Errorf("expected %s: %v, got: %v", CorsAccessControlAllowMethodsHeaderName, expectedAllowedMethods, actual)
	}

	expectedAllowedHosts := "https://example.org"
	actual = recorder.Header().Get(CorsAccessControlAllowOriginHeaderName)
	if actual != expectedAllowedHosts {
		t.Errorf("expected %s: %v, got: %v", CorsAccessControlAllowOriginHeaderName, expectedAllowedHosts, actual)
	}
}

func TestAddCorsHeaders_NeitherRequestHeader_Success(t *testing.T) {
	middleware := NewMiddleware([]string{"example.org"}, nil)

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodOptions, "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Add("FooBar", "https://example.org")

	middleware.ServeHTTP(recorder, request)

	expectedAllowedMethods := "GET, POST"
	actual := recorder.Header().Get(CorsAccessControlAllowMethodsHeaderName)
	if actual != expectedAllowedMethods {
		t.Errorf("expected %s: %v, got: %v", CorsAccessControlAllowMethodsHeaderName, expectedAllowedMethods, actual)
	}

	expectedAllowedHosts := ""
	actual = recorder.Header().Get(CorsAccessControlAllowOriginHeaderName)
	if actual != expectedAllowedHosts {
		t.Errorf("expected %s: %v, got: %v", CorsAccessControlAllowOriginHeaderName, expectedAllowedHosts, actual)
	}
}

func TestAddCorsHeaders_UnparseableRequestHeader_Success(t *testing.T) {
	middleware := NewMiddleware([]string{"example.org"}, nil)

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodOptions, "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Add("Origin", "foo-bar-bang:/\bingboom")

	middleware.ServeHTTP(recorder, request)

	expectedAllowedMethods := "GET, POST"
	actual := recorder.Header().Get(CorsAccessControlAllowMethodsHeaderName)
	if actual != expectedAllowedMethods {
		t.Errorf("expected %s: %v, got: %v", CorsAccessControlAllowMethodsHeaderName, expectedAllowedMethods, actual)
	}

	expectedAllowedHosts := ""
	actual = recorder.Header().Get(CorsAccessControlAllowOriginHeaderName)
	if actual != expectedAllowedHosts {
		t.Errorf("expected %s: %v, got: %v", CorsAccessControlAllowOriginHeaderName, expectedAllowedHosts, actual)
	}
}
