package auth

import (
	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/auth/libp2p"
	"go.uber.org/fx"
)

var Module = fx.Options(
	libp2p.Module,
)
