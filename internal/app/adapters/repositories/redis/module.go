package redis

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewRedisClient),
	fx.Provide(NewNonceCache),
	fx.Provide(NewLeaseCache),
)
