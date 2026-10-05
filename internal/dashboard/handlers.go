package dashboard

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"time"
)

//go:embed static
var staticFS embed.FS

type Source interface {
	Records(ctx context.Context) ([]json.RawMessage, error)
}

type Server struct {
	db     *sql.DB
	source Source
	mux    *http.ServeMux
}

func New(db *sql.DB, source Source) *Server {
	s := &Server{db: db, source: source, mux: http.NewServeMux()}
	s.mux.HandleFunc("GET /healthz", s.health)
	s.mux.HandleFunc("GET /api/records", s.listRecords)
	s.mux.HandleFunc("POST /api/refresh", s.refresh)
	assets, _ := fs.Sub(staticFS, "static")
	s.mux.Handle("/", http.FileServer(http.FS(assets)))
	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("ok"))
}

func (s *Server) listRecords(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `SELECT payload FROM records ORDER BY id`)
	if err != nil {
		http.Error(w, "read records", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := make([]json.RawMessage, 0)
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			http.Error(w, "read records", http.StatusInternalServerError)
			return
		}
		out = append(out, json.RawMessage(payload))
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "read records", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"records": out})
}

func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	if s.source == nil {
		http.Error(w, "no data source configured", http.StatusServiceUnavailable)
		return
	}
	records, err := s.source.Records(r.Context())
	if err != nil {
		// Never log the raw error: a source URL may carry credentials in its
		// query string, and an integration could include that text. Log only the
		// error type as a safe diagnostic category.
		log.Printf("refresh failed (%T)", err)
		http.Error(w, "data source refresh failed", http.StatusBadGateway)
		return
	}
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "store records", http.StatusInternalServerError)
		return
	}
	if _, err := tx.ExecContext(r.Context(), `DELETE FROM records`); err != nil {
		tx.Rollback()
		http.Error(w, "store records", http.StatusInternalServerError)
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	for _, record := range records {
		if _, err := tx.ExecContext(r.Context(), `INSERT INTO records(payload, fetched_at) VALUES(?, ?)`, string(record), now); err != nil {
			tx.Rollback()
			http.Error(w, "store records", http.StatusInternalServerError)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, "store records", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"count": len(records), "fetched_at": now})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
