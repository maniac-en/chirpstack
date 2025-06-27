package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPlatform_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		want     bool
	}{
		{
			name:     "valid dev platform",
			platform: PlatformDev,
			want:     true,
		},
		{
			name:     "valid prod platform",
			platform: PlatformProd,
			want:     true,
		},
		{
			name:     "invalid platform",
			platform: Platform("invalid"),
			want:     false,
		},
		{
			name:     "empty platform",
			platform: Platform(""),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.platform.IsValid(); got != tt.want {
				t.Errorf("Platform.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidPlatforms(t *testing.T) {
	platforms := ValidPlatforms()

	expected := []Platform{PlatformDev, PlatformProd}
	if len(platforms) != len(expected) {
		t.Errorf("ValidPlatforms() length = %v, want %v", len(platforms), len(expected))
	}

	for i, platform := range platforms {
		if platform != expected[i] {
			t.Errorf("ValidPlatforms()[%d] = %v, want %v", i, platform, expected[i])
		}
	}
}

func TestParsePlatform(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Platform
		wantErr bool
	}{
		{
			name:    "valid dev platform",
			input:   "dev",
			want:    PlatformDev,
			wantErr: false,
		},
		{
			name:    "valid prod platform",
			input:   "prod",
			want:    PlatformProd,
			wantErr: false,
		},
		{
			name:    "invalid platform",
			input:   "staging",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "case sensitive",
			input:   "DEV",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePlatform(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePlatform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParsePlatform() = %v, want %v", got, tt.want)
			}

			if tt.wantErr && err != nil {
				expectedErr := "invalid platform '" + tt.input + "', must be one of: [dev prod]"
				if err.Error() != expectedErr {
					t.Errorf("ParsePlatform() error message = %v, want %v", err.Error(), expectedErr)
				}
			}
		})
	}
}

func TestAPIConfig_MiddlewareMetricsInc(t *testing.T) {
	cfg := &APIConfig{}

	// Create a simple handler that we'll wrap
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	// Wrap the handler with metrics middleware
	wrappedHandler := cfg.MiddlewareMetricsInc(handler)

	// Test initial state
	if cfg.fileserverHits.Load() != 0 {
		t.Errorf("Initial fileserverHits = %v, want 0", cfg.fileserverHits.Load())
	}

	// Make first request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	// Check that hits were incremented
	if cfg.fileserverHits.Load() != 1 {
		t.Errorf("After first request fileserverHits = %v, want 1", cfg.fileserverHits.Load())
	}

	// Check that Cache-Control header was set
	if w.Header().Get("Cache-Control") != "no-cache" {
		t.Errorf("Cache-Control header = %v, want 'no-cache'", w.Header().Get("Cache-Control"))
	}

	// Check that original handler was called
	if w.Code != http.StatusOK {
		t.Errorf("Response code = %v, want %v", w.Code, http.StatusOK)
	}

	if w.Body.String() != "test" {
		t.Errorf("Response body = %v, want 'test'", w.Body.String())
	}

	// Make second request
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w2, req2)

	// Check that hits were incremented again
	if cfg.fileserverHits.Load() != 2 {
		t.Errorf("After second request fileserverHits = %v, want 2", cfg.fileserverHits.Load())
	}
}

func TestAPIConfig_LogMiddleware(t *testing.T) {
	cfg := &APIConfig{}

	// Create a simple handler that we'll wrap
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	// Wrap the handler with log middleware
	wrappedHandler := cfg.LogMiddleware(handler)

	// Make request
	req := httptest.NewRequest("POST", "/api/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	// Check that original handler was called
	if w.Code != http.StatusOK {
		t.Errorf("Response code = %v, want %v", w.Code, http.StatusOK)
	}

	if w.Body.String() != "test" {
		t.Errorf("Response body = %v, want 'test'", w.Body.String())
	}

	// Note: We can't easily test the logging output without capturing log output,
	// but we can verify the handler chain works correctly
}

func TestAPIConfig_Concurrent_MiddlewareMetricsInc(t *testing.T) {
	cfg := &APIConfig{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := cfg.MiddlewareMetricsInc(handler)

	// Run multiple concurrent requests
	const numRequests = 10
	for range numRequests {
		go func() {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w, req)
		}()
	}
}

func TestAPIConfig_MiddlewareChaining(t *testing.T) {
	cfg := &APIConfig{}

	// Create a handler that records what headers it sees
	var receivedHeaders http.Header
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	})

	// Chain both middlewares
	wrappedHandler := cfg.LogMiddleware(cfg.MiddlewareMetricsInc(handler))

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	// Check that both middlewares worked
	if cfg.fileserverHits.Load() != 1 {
		t.Errorf("fileserverHits = %v, want 1", cfg.fileserverHits.Load())
	}

	if w.Header().Get("Cache-Control") != "no-cache" {
		t.Errorf("Cache-Control header = %v, want 'no-cache'", w.Header().Get("Cache-Control"))
	}

	// Check that original request headers were preserved
	if receivedHeaders.Get("User-Agent") != "test-agent" {
		t.Errorf("User-Agent header = %v, want 'test-agent'", receivedHeaders.Get("User-Agent"))
	}
}

func TestAPIConfig_RefreshUserToken_MissingAuthHeader(t *testing.T) {
	cfg := &APIConfig{}

	req := httptest.NewRequest("POST", "/api/refresh", nil)
	w := httptest.NewRecorder()

	cfg.RefreshUserToken(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	expected := `{"error":"authorization header not found in request"}`
	if strings.TrimSpace(w.Body.String()) != expected {
		t.Errorf("Expected body %q, got %q", expected, strings.TrimSpace(w.Body.String()))
	}
}

func TestAPIConfig_RefreshUserToken_InvalidAuthHeader(t *testing.T) {
	cfg := &APIConfig{}

	req := httptest.NewRequest("POST", "/api/refresh", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()

	cfg.RefreshUserToken(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	expected := `{"error":"invalid authorization header found in request"}`
	if strings.TrimSpace(w.Body.String()) != expected {
		t.Errorf("Expected body %q, got %q", expected, strings.TrimSpace(w.Body.String()))
	}
}

