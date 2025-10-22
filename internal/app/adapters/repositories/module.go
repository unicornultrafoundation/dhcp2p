package repositories

import (
	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/repositories/hybrid"
	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/repositories/postgres"
	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/repositories/redis"
	"go.uber.org/fx"
)

var Module = fx.Options(
	// ethereum.Module, // Will use later for signing IP allocations
	postgres.Module,
	redis.Module,
	hybrid.Module,
)
