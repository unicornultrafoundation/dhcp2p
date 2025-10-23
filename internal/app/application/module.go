package application

import (
	"github.com/unicornultrafoundation/dhcp2p/internal/app/application/jobs"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/application/services"
	"go.uber.org/fx"
)

var Module = fx.Options(
	services.Module,
	jobs.Module,
)
