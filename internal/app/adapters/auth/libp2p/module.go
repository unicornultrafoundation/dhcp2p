package libp2p

import (
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/ports"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewSignatureVerifier,
			fx.As(new(ports.SignatureVerifier)),
		),
	),
)
