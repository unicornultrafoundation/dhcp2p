package ports

import (
	"context"

	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/models"
)

type AuthService interface {
	RequestAuth(ctx context.Context, request *models.AuthRequest) (*models.AuthResponse, error)
	VerifyAuth(ctx context.Context, request *models.AuthVerifyRequest) (*models.AuthVerifyResponse, error)
}
