package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	sessionsvc "ndbx-app/internal/service/session"
)

// Handler contains HTTP endpoints.
type Handler struct {
	sessions   *sessionsvc.Service
	ttlSeconds int
	timeout    time.Duration
}

// NewHandler creates handler with injected dependencies.
func NewHandler(sessions *sessionsvc.Service, ttlSeconds int) *Handler {
	return &Handler{
		sessions:   sessions,
		ttlSeconds: ttlSeconds,
		timeout:    3 * time.Second,
	}
}

// Register mounts all application routes.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/session", h.handleSession)
	mux.HandleFunc("/openapi.yaml", h.handleOpenAPI)
	mux.HandleFunc("/swagger", h.handleSwaggerUI)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if sid := readSIDCookie(r); sid != "" {
		writeSIDCookie(w, sid, h.ttlSeconds)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("failed to encode health response: %v", err)
	}
}

func (h *Handler) handleSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	result, err := h.sessions.UpsertSession(ctx, readSIDCookie(r))
	if err != nil {
		log.Printf("failed to upsert session: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeSIDCookie(w, result.SID, h.ttlSeconds)
	if result.Created {
		w.WriteHeader(http.StatusCreated)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func readSIDCookie(r *http.Request) string {
	ck, err := r.Cookie(sessionsvc.CookieName)
	if err != nil {
		return ""
	}
	if !sessionsvc.IsValidSID(ck.Value) {
		return ""
	}
	return ck.Value
}

func writeSIDCookie(w http.ResponseWriter, sid string, ttlSeconds int) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionsvc.CookieName,
		Value:    sid,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   ttlSeconds,
	})
}
