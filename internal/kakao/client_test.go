package kakao

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/qj0r9j0vc2/kko/internal/config"
)

func testClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	cfg := &config.Config{APIKey: "test-api-key"}
	c := NewClient(cfg)
	return c, srv
}

func TestClient_Get_APIKeyAuth(t *testing.T) {
	var gotAuth string
	c, srv := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer srv.Close()

	_, err := c.Get(context.Background(), srv.URL+"/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "KakaoAK test-api-key" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "KakaoAK test-api-key")
	}
}

func TestClient_Get_NoAPIKey(t *testing.T) {
	cfg := &config.Config{APIKey: ""}
	c := NewClient(cfg)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	_, err := c.Get(context.Background(), srv.URL+"/test", nil)
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
}

func TestClient_RetryOn429(t *testing.T) {
	var attempts int32
	c, srv := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(APIError{Code: -10, Message: "rate limited"})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
	}))
	defer srv.Close()

	_, err := c.Get(context.Background(), srv.URL+"/test", nil)
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClient_NoRetryOn400(t *testing.T) {
	var attempts int32
	c, srv := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIError{Code: -2, Message: "bad request"})
	}))
	defer srv.Close()

	_, err := c.Get(context.Background(), srv.URL+"/test", nil)
	if err == nil {
		t.Fatal("expected error on 400")
	}
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("expected 1 attempt (no retry on 4xx), got %d", attempts)
	}
}

func TestClient_RetryOn500(t *testing.T) {
	var attempts int32
	c, srv := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(APIError{Code: -530, Message: "internal error"})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
	}))
	defer srv.Close()

	_, err := c.Get(context.Background(), srv.URL+"/test", nil)
	if err != nil {
		t.Fatalf("expected success after 500 retry, got %v", err)
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}
