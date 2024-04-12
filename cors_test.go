package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddCorsHeaders_OriginRequestHeader_Success(t *testing.T) {
	*hostAllowlist = "example.org"

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodOptions, "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Add("Origin", "https://example.org")

	addCorsHeaders(recorder, request)

	expectedAllowedMethods := "GET"
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
	*hostAllowlist = "example.org"

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodOptions, "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Add("Referer", "https://example.org")

	addCorsHeaders(recorder, request)

	expectedAllowedMethods := "GET"
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
	*hostAllowlist = "example.org"

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodOptions, "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Add("FooBar", "https://example.org")

	addCorsHeaders(recorder, request)

	expectedAllowedMethods := "GET"
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
	*hostAllowlist = "example.org"

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodOptions, "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Add("Origin", "foo-bar-bang:/\bingboom")

	addCorsHeaders(recorder, request)

	expectedAllowedMethods := "GET"
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
