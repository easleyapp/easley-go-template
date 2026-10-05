package dashboard

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"easley.app/dashboard-starter/internal/db"
)

type fakeSource struct {
	records []json.RawMessage
	err     error
}

func (f fakeSource) Records(context.Context) ([]json.RawMessage, error) {
	return f.records, f.err
}

func newTestServer(t *testing.T, source Source) *Server {
	t.Helper()
	database, err := db.Open(filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return New(database, source)
}

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestServer(t, nil).Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != "ok" {
		t.Fatalf("healthz = %d %q", rec.Code, rec.Body.String())
	}
}

func TestRefreshStoresAndLists(t *testing.T) {
	src := fakeSource{records: []json.RawMessage{
		json.RawMessage(`{"amount":10}`),
		json.RawMessage(`{"amount":20}`),
	}}
	srv := newTestServer(t, src)

	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/refresh", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh = %d %q", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/records", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("records = %d %q", rec.Code, rec.Body.String())
	}
	var body struct {
		Records []map[string]int `json:"records"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Records) != 2 || body.Records[0]["amount"] != 10 {
		t.Fatalf("unexpected records: %+v", body.Records)
	}
}

func TestRefreshWithoutSource(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestServer(t, nil).Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/refresh", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("refresh without source = %d", rec.Code)
	}
}

func TestRefreshErrorDoesNotLeak(t *testing.T) {
	var logs bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(previous) })

	src := fakeSource{err: errors.New("boom http://source.example/data?token=supersecret")}
	rec := httptest.NewRecorder()
	newTestServer(t, src).Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/refresh", nil))
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("refresh error = %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "supersecret") {
		t.Fatalf("response leaked source details: %q", rec.Body.String())
	}
	if strings.Contains(logs.String(), "supersecret") {
		t.Fatalf("logs leaked source details: %q", logs.String())
	}
}

func TestRecordsPersistAcrossRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.db")

	first, err := db.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Migrate(first); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	rec := httptest.NewRecorder()
	New(first, fakeSource{records: []json.RawMessage{json.RawMessage(`{"a":1}`)}}).
		Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/refresh", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh = %d", rec.Code)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	reopened, err := db.Open(path)
	if err != nil {
		t.Fatalf("reopen db: %v", err)
	}
	defer reopened.Close()
	if err := db.Migrate(reopened); err != nil {
		t.Fatalf("migrate after restart: %v", err)
	}
	rec = httptest.NewRecorder()
	New(reopened, nil).Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/records", nil))
	if !strings.Contains(rec.Body.String(), `"a":1`) {
		t.Fatalf("records not persisted across restart: %s", rec.Body.String())
	}
}
