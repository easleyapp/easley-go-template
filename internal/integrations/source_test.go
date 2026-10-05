package integrations

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecordsStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	h := NewHTTP(srv.URL+"/data?token=supersecret", "key")
	_, err := h.Records(context.Background())
	if err == nil {
		t.Fatal("expected an error")
	}
	if strings.Contains(err.Error(), "supersecret") || strings.Contains(err.Error(), srv.URL) {
		t.Fatalf("error leaked the source URL: %v", err)
	}
}

func TestRecordsTransportError(t *testing.T) {
	h := NewHTTP("http://127.0.0.1:1/data?token=supersecret", "")
	_, err := h.Records(context.Background())
	if err == nil {
		t.Fatal("expected an error")
	}
	if strings.Contains(err.Error(), "supersecret") || strings.Contains(err.Error(), "127.0.0.1:1") {
		t.Fatalf("error leaked the source URL: %v", err)
	}
}

func TestRecordsSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`[{"amount":10}]`))
	}))
	defer srv.Close()

	h := NewHTTP(srv.URL, "")
	records, err := h.Records(context.Background())
	if err != nil {
		t.Fatalf("Records: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1", len(records))
	}
}
