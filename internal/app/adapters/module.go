package adapters

import (
	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/handlers"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/repositories"
	"go.uber.org/fx"
)

var Module = fx.Options(
	handlers.Module,
	repositories.Module,
)
