package infrastructure

import (
	"github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure/config"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure/logger"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	config.Module,
	logger.Module,
	server.Module,
)
