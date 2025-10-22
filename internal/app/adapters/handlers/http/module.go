package http

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewLeaseHandler),
	fx.Provide(NewAuthHandler),
	fx.Provide(NewHealthHandler),
	fx.Provide(NewHTTPRouter),
)
