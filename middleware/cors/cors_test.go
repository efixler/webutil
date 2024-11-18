package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCORSOptions(t *testing.T) {
	test := []struct {
		name          string
		opts          []corsOption
		expectOrigin  string
		expectMethods string
		expectHeaders string
		expectMaxAge  string
	}{
		{
			name:          "default",
			expectOrigin:  "*",
			expectMethods: "*",
			expectHeaders: "*",
			expectMaxAge:  "86400",
		},
		{
			name:          "allow origins",
			opts:          []corsOption{AllowOrigin("http://example.com")},
			expectOrigin:  "http://example.com",
			expectMethods: "*",
			expectHeaders: "*",
			expectMaxAge:  "86400",
		},
		{
			name:          "allow methods",
			opts:          []corsOption{AllowMethods(http.MethodGet, http.MethodPost)},
			expectOrigin:  "*",
			expectMethods: "GET,POST",
			expectHeaders: "*",
			expectMaxAge:  "86400",
		},
		{
			name:          "allow headers",
			opts:          []corsOption{AllowHeaders("X-Custom-Header")},
			expectOrigin:  "*",
			expectMethods: "*",
			expectHeaders: "X-Custom-Header",
			expectMaxAge:  "86400",
		},
		{
			name:          "max age",
			opts:          []corsOption{MaxAge(1 * time.Second)},
			expectOrigin:  "*",
			expectMethods: "*",
			expectHeaders: "*",
			expectMaxAge:  "1",
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			handlerWasRun := false
			h := func(w http.ResponseWriter, r *http.Request) {
				handlerWasRun = true
			}
			cors := CORS(tt.opts...)
			handler := cors(h)
			if handler == nil {
				t.Fatal("handler is nil")
			}
			req := httptest.NewRequest(http.MethodOptions, "http://example.com/foo", nil)
			w := httptest.NewRecorder()
			handler(w, req)
			if handlerWasRun {
				t.Error("Chain did not terminate for OPTIONS request")
			}
			resp := w.Result()
			if resp.Header.Get("Access-Control-Allow-Origin") != tt.expectOrigin {
				t.Errorf("unexpected Access-Control-Allow-Origin: got %s, want %s", resp.Header.Get("Access-Control-Allow-Origin"), tt.expectOrigin)
			}
			if resp.Header.Get("Access-Control-Allow-Methods") != tt.expectMethods {
				t.Errorf("unexpected Access-Control-Allow-Methods: got %s, want %s", resp.Header.Get("Access-Control-Allow-Methods"), tt.expectMethods)
			}
			if resp.Header.Get("Access-Control-Allow-Headers") != tt.expectHeaders {
				t.Errorf("unexpected Access-Control-Allow-Headers: got %s, want %s", resp.Header.Get("Access-Control-Allow-Headers"), tt.expectHeaders)
			}
			if resp.Header.Get("Access-Control-Max-Age") != tt.expectMaxAge {
				t.Errorf("unexpected Access-Control-Max-Age: got %s, want %s", resp.Header.Get("Access-Control-Max-Age"), tt.expectMaxAge)
			}
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		opts         []corsOption
		expectOrigin string
	}{
		{
			name:         "default",
			expectOrigin: "*",
		},
		{
			name:         "allow specific origin",
			opts:         []corsOption{AllowOrigin("http://example.com")},
			expectOrigin: "http://example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerWasRun := false
			h := func(w http.ResponseWriter, r *http.Request) {
				handlerWasRun = true
			}
			cors := CORS(tt.opts...)
			handler := cors(h)
			if handler == nil {
				t.Fatal("handler is nil")
			}
			req := httptest.NewRequest(http.MethodGet, "http://example.com/foo", nil)
			w := httptest.NewRecorder()
			handler(w, req)
			if !handlerWasRun {
				t.Error("handler was not run")
			}
			resp := w.Result()
			if resp.Header.Get("Access-Control-Allow-Origin") != tt.expectOrigin {
				t.Errorf("unexpected Access-Control-Allow-Origin: got %s, want *", resp.Header.Get("Access-Control-Allow-Origin"))
			}
			if resp.Header.Get("Access-Control-Allow-Methods") != "" {
				t.Errorf("unexpected Access-Control-Allow-Methods: got %s, want empty", resp.Header.Get("Access-Control-Allow-Methods"))
			}
			if resp.Header.Get("Access-Control-Allow-Headers") != "" {
				t.Errorf("unexpected Access-Control-Allow-Headers: got %s, want empty", resp.Header.Get("Access-Control-Allow-Headers"))
			}
			if resp.Header.Get("Access-Control-Max-Age") != "" {
				t.Errorf("unexpected Access-Control-Max-Age: got %s, want empty", resp.Header.Get("Access-Control-Max-Age"))
			}
		})
	}
}
