package services

import (
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/ports"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewNonceService,
			fx.As(new(ports.NonceService)),
		),
		fx.Annotate(
			NewLeaseService,
			fx.As(new(ports.LeaseService)),
		),
		fx.Annotate(
			NewAuthService,
			fx.As(new(ports.AuthService)),
		),
	),
)
