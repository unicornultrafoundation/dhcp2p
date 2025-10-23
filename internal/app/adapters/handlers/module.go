package handlers

import (
	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/auth"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/handlers/http"
	"go.uber.org/fx"
)

var Module = fx.Options(
	http.Module,
	auth.Module,
)
