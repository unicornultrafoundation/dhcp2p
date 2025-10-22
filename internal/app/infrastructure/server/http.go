package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	handlers "github.com/duchuongnguyen/dhcp2p/internal/app/adapters/handlers/http"
	"github.com/duchuongnguyen/dhcp2p/internal/app/infrastructure/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type HTTPServer struct {
	server *http.Server
}

func NewHTTPServer(lc fx.Lifecycle, cfg *config.AppConfig, router *handlers.Router, logger *zap.Logger) *HTTPServer {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router.Mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
			if err != nil {
				return err
			}

			logger.With(zap.Int("port", cfg.Port)).Info("HTTPServer is running")

			go server.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})

	return &HTTPServer{
		server: server,
	}
}
