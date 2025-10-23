package app

import (
	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/application"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/ports"
	infrastructure "github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure/server"
	"go.uber.org/fx"
)

func NewApp() *fx.App {
	return fx.New(
		// Use the zap logger (for debugging)
		// fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
		// 	return &fxevent.ZapLogger{Logger: log}
		// }),

		// No logger for now
		fx.NopLogger,

		// Add modules
		adapters.Module,
		application.Module,
		infrastructure.Module,

		// Invoke the servers
		fx.Invoke(func(server *server.HTTPServer) {}),

		// Invoke the jobs
		fx.Invoke(func(nonceCleaner ports.NonceCleaner) {}),
	)
}
