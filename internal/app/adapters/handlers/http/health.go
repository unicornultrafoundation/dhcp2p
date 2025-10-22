package http

import (
	"context"
	"net/http"
	"time"

	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	db    *pgxpool.Pool
	cache *redis.Client
}

func NewHealthHandler(db *pgxpool.Pool, cache *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, cache: cache}
}

// Health is a lightweight liveness check
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Readiness verifies dependencies (DB and Cache)
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	if h.db == nil || h.cache == nil {
		writeErrorResponse(w, http.StatusServiceUnavailable, errors.ErrMissingDependencies)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	// Check DB
	if err := h.db.Ping(ctx); err != nil {
		writeErrorResponse(w, http.StatusServiceUnavailable, err)
		return
	}

	// Check Cache
	if err := h.cache.Ping(ctx).Err(); err != nil {
		writeErrorResponse(w, http.StatusServiceUnavailable, err)
		return
	}

	writeResponse(w, http.StatusOK, map[string]string{"status": "ready"})
}
