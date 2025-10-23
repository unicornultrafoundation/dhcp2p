package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/application/services"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/models"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure/config"
	"github.com/unicornultrafoundation/dhcp2p/tests/mocks"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestLeaseService_AllocateIP(t *testing.T) {
	tests := []struct {
		name           string
		peerID         string
		mockSetup      func(*gomock.Controller, *mocks.MockLeaseRepository)
		expectedResult *models.Lease
		expectedError  error
	}{
		{
			name:   "successful allocation - new lease",
			peerID: "peer123",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockLeaseRepository) {
				mockRepo.EXPECT().GetLeaseByPeerID(gomock.Any(), "peer123").Return(nil, nil)
				mockRepo.EXPECT().FindAndReuseExpiredLease(gomock.Any(), "peer123").Return(nil, nil).AnyTimes()
				mockRepo.EXPECT().AllocateNewLease(gomock.Any(), "peer123").Return(&models.Lease{
					TokenID:   167772161,
					PeerID:    "peer123",
					CreatedAt: time.Now(),
					ExpiresAt: time.Now().Add(time.Hour),
				}, nil)
			},
			expectedResult: &models.Lease{
				TokenID: 167772161,
				PeerID:  "peer123",
			},
			expectedError: nil,
		},
		{
			name:   "successful allocation - existing lease",
			peerID: "peer456",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockLeaseRepository) {
				mockRepo.EXPECT().GetLeaseByPeerID(gomock.Any(), "peer456").Return(&models.Lease{
					TokenID:   167772162,
					PeerID:    "peer456",
					CreatedAt: time.Now(),
					ExpiresAt: time.Now().Add(time.Hour),
				}, nil)
			},
			expectedResult: &models.Lease{
				TokenID: 167772162,
				PeerID:  "peer456",
			},
			expectedError: nil,
		},
		{
			name:   "successful allocation - reuse expired lease",
			peerID: "peer789",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockLeaseRepository) {
				mockRepo.EXPECT().GetLeaseByPeerID(gomock.Any(), "peer789").Return(nil, nil)
				mockRepo.EXPECT().FindAndReuseExpiredLease(gomock.Any(), "peer789").Return(&models.Lease{
					TokenID:   167772163,
					PeerID:    "peer789",
					CreatedAt: time.Now(),
					ExpiresAt: time.Now().Add(time.Hour),
				}, nil)
			},
			expectedResult: &models.Lease{
				TokenID: 167772163,
				PeerID:  "peer789",
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockLeaseRepository(ctrl)
			tt.mockSetup(ctrl, mockRepo)

			service := services.NewLeaseService(&config.AppConfig{
				MaxLeaseRetries: 3,
				LeaseRetryDelay: 100,
			}, mockRepo, zap.NewNop())

			result, err := service.AllocateIP(context.Background(), tt.peerID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.TokenID, result.TokenID)
				assert.Equal(t, tt.expectedResult.PeerID, result.PeerID)
			}
		})
	}
}

func TestLeaseService_GetLeaseByPeerID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockLeaseRepository(ctrl)
	service := services.NewLeaseService(&config.AppConfig{}, mockRepo, zap.NewNop())

	expectedLease := &models.Lease{
		TokenID:   167772161,
		PeerID:    "peer123",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	mockRepo.EXPECT().GetLeaseByPeerID(gomock.Any(), "peer123").Return(expectedLease, nil)

	result, err := service.GetLeaseByPeerID(context.Background(), "peer123")

	assert.NoError(t, err)
	assert.Equal(t, expectedLease, result)
}

func TestLeaseService_GetLeaseByTokenID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockLeaseRepository(ctrl)
	service := services.NewLeaseService(&config.AppConfig{}, mockRepo, zap.NewNop())

	expectedLease := &models.Lease{
		TokenID:   167772161,
		PeerID:    "peer123",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	mockRepo.EXPECT().GetLeaseByTokenID(gomock.Any(), int64(167772161)).Return(expectedLease, nil)

	result, err := service.GetLeaseByTokenID(context.Background(), 167772161)

	assert.NoError(t, err)
	assert.Equal(t, expectedLease, result)
}

func TestLeaseService_RenewLease(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockLeaseRepository(ctrl)
	service := services.NewLeaseService(&config.AppConfig{}, mockRepo, zap.NewNop())

	expectedLease := &models.Lease{
		TokenID:   167772161,
		PeerID:    "peer123",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	mockRepo.EXPECT().RenewLease(gomock.Any(), int64(167772161), "peer123").Return(expectedLease, nil)

	result, err := service.RenewLease(context.Background(), 167772161, "peer123")

	assert.NoError(t, err)
	assert.Equal(t, expectedLease, result)
}

func TestLeaseService_ReleaseLease(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockLeaseRepository(ctrl)
	service := services.NewLeaseService(&config.AppConfig{}, mockRepo, zap.NewNop())

	mockRepo.EXPECT().ReleaseLease(gomock.Any(), int64(167772161), "peer123").Return(nil)

	err := service.ReleaseLease(context.Background(), 167772161, "peer123")

	assert.NoError(t, err)
}
