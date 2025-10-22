package postgres

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	// Repositories
	fx.Provide(NewDBPool),
	fx.Provide(NewNonceRepository),
	fx.Provide(NewLeaseRepository),
)
