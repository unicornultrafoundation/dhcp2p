package jobs

import (
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/ports"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(NewNonceCleanerJob, fx.As(new(ports.NonceCleaner))),
	),
)
