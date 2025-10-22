package hybrid

import (
	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/repositories/postgres"
	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/repositories/redis"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/ports"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Options(
	fx.Provide(
		// Wrap DB repos with caches to expose as default implementations
		fx.Annotate(
			func(
				logger *zap.Logger,
				dbNonceRepo *postgres.NonceRepository,
				cache *redis.NonceCache,
			) ports.NonceRepository {
				return NewNonceRepository(dbNonceRepo, cache, logger)
			},
			fx.As(new(ports.NonceRepository)),
		),
		fx.Annotate(
			func(
				logger *zap.Logger,
				dbLeaseRepo *postgres.LeaseRepository,
				cache *redis.LeaseCache,
			) ports.LeaseRepository {
				return NewLeaseRepository(dbLeaseRepo, cache, logger)
			},
			fx.As(new(ports.LeaseRepository)),
		),
	),
)
