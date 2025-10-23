package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/handlers/http/utils"
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
	utils.WriteResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Readiness verifies dependencies (DB and Cache)
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	if h.db == nil || h.cache == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"type":    "internal_error",
			"code":    "MISSING_DEPENDENCIES",
			"message": "Missing required dependencies",
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	// Check DB
	if err := h.db.Ping(ctx); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"type":    "internal_error",
			"code":    "DATABASE_CHECK_FAILED",
			"message": "Database health check failed",
		})
		return
	}

	// Check Cache
	if err := h.cache.Ping(ctx).Err(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"type":    "internal_error",
			"code":    "REDIS_CHECK_FAILED",
			"message": "Redis health check failed",
		})
		return
	}

	utils.WriteResponse(w, http.StatusOK, map[string]string{"status": "ready"})
}
