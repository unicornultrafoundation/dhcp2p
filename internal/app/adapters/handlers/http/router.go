package http

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	httpMiddleware "github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/handlers/http/middleware"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure/config"
)

type Router struct {
	*chi.Mux
}

func NewHTTPRouter(logger *zap.Logger, authHandler *AuthHandler, leaseHandler *LeaseHandler, healthHandler *HealthHandler, cfg *config.AppConfig) *Router {
	r := chi.NewRouter()

	// Apply security middleware to all routes
	r.Use(httpMiddleware.CombinedSecurityMiddleware())

	// Apply IP-based rate limiting
	r.Use(httpMiddleware.RateLimitMiddleware(cfg, logger))

	// Apply standard middleware
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: zap.NewStdLog(logger), NoColor: false}))
	r.Use(middleware.Recoverer)                 // recover from panics
	r.Use(middleware.Timeout(60 * time.Second)) // set timeout

	// Protected routes
	r.Group(func(pr chi.Router) {
		// Authentication middleware
		pr.Use(
			httpMiddleware.WithAuth(authHandler.authService),
		)

		// Lease routes
		pr.Post("/allocate-ip", leaseHandler.AllocateIP)
		pr.Post("/renew-lease", leaseHandler.RenewLease)
		pr.Post("/release-lease", leaseHandler.ReleaseLease)
	})

	// Public routes
	r.Get("/lease/peer-id/{peerID}", leaseHandler.GetLeaseByPeerID)
	r.Get("/lease/token-id/{tokenID}", leaseHandler.GetLeaseByTokenID)

	// Auth routes
	r.Post("/request-auth", authHandler.RequestAuth)

	// Health check routes (no authentication required)
	r.Get("/health", healthHandler.Health)
	r.Get("/ready", healthHandler.Readiness)

	return &Router{
		Mux: r,
	}
}
