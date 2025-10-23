package repositories

import (
	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/repositories/hybrid"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/repositories/postgres"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/repositories/redis"
	"go.uber.org/fx"
)

var Module = fx.Options(
	postgres.Module,
	redis.Module,
	hybrid.Module,
)
