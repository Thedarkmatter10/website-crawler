package crawler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckWebsite(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.Write([]byte("User-agent: *\nAllow: /"))
			return
		}
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	tests := []struct {
		name        string
		url         string
		wantStatus  int
		wantTimeMin int64
		wantTimeMax int64
	}{
		{
			name:        "Valid URL",
			url:         server.URL,
			wantStatus:  200,
			wantTimeMin: 0,
			wantTimeMax: 1000,
		},
		{
			name:        "Invalid URL",
			url:         "http://nonexistent.invalid",
			wantStatus:  0,
			wantTimeMin: 0,
			wantTimeMax: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, responseTime := CheckWebsite(tt.url)
			if status != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, status)
			}
			if responseTime < tt.wantTimeMin || responseTime > tt.wantTimeMax {
				t.Errorf("Response time %d out of range [%d, %d]", responseTime, tt.wantTimeMin, tt.wantTimeMax)
			}
		})
	}
}
