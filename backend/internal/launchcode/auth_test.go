package launchcode

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIAuthMiddlewareDisabledAllowsAPIRequest(t *testing.T) {
	server := &LaunchServer{}
	server.SetAPIAuthConfig(APIAuthConfig{Enabled: false})

	status, called := invokeAuthMiddleware(server, "/api/profiles", "")
	if status != http.StatusNoContent || !called {
		t.Fatalf("disabled auth should pass through: status=%d called=%v", status, called)
	}
}

func TestAPIAuthMiddlewareEnabledWithoutKeyFailsClosed(t *testing.T) {
	server := &LaunchServer{}
	server.SetAPIAuthConfig(APIAuthConfig{Enabled: true})

	status, called := invokeAuthMiddleware(server, "/api/profiles", "")
	if status != http.StatusServiceUnavailable || called {
		t.Fatalf("unconfigured auth should fail closed: status=%d called=%v", status, called)
	}
}

func TestAPIAuthMiddlewareRejectsInvalidKey(t *testing.T) {
	server := &LaunchServer{}
	server.SetAPIAuthConfig(APIAuthConfig{Enabled: true, APIKey: "expected-key"})

	status, called := invokeAuthMiddleware(server, "/api/profiles", "wrong-key")
	if status != http.StatusUnauthorized || called {
		t.Fatalf("invalid key should be rejected: status=%d called=%v", status, called)
	}
}

func TestAPIAuthMiddlewareAcceptsValidKey(t *testing.T) {
	server := &LaunchServer{}
	server.SetAPIAuthConfig(APIAuthConfig{Enabled: true, APIKey: "expected-key"})

	status, called := invokeAuthMiddleware(server, "/api/profiles", "expected-key")
	if status != http.StatusNoContent || !called {
		t.Fatalf("valid key should pass through: status=%d called=%v", status, called)
	}
}

func TestAPIAuthMiddlewareLeavesNonAPIEndpointAvailable(t *testing.T) {
	server := &LaunchServer{}
	server.SetAPIAuthConfig(APIAuthConfig{Enabled: true})

	status, called := invokeAuthMiddleware(server, "/health", "")
	if status != http.StatusNoContent || !called {
		t.Fatalf("non-api endpoint should pass through: status=%d called=%v", status, called)
	}
}

func invokeAuthMiddleware(server *LaunchServer, path, apiKey string) (int, bool) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	if apiKey != "" {
		request.Header.Set(DefaultAPIKeyHeader, apiKey)
	}

	server.apiAuthMiddleware(next).ServeHTTP(recorder, request)
	return recorder.Code, called
}
