package hybrid

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duchuongnguyen/dhcp2p/internal/app/adapters/repositories/hybrid"
	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/models"
	"github.com/duchuongnguyen/dhcp2p/tests/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestLeaseRepository_GetLeaseByPeerID(t *testing.T) {
	tests := []struct {
		name          string
		peerID        string
		mockSetup     func(*gomock.Controller, *mocks.MockLeaseRepository, *mocks.MockLeaseCache)
		expectedLease *models.Lease
		expectedError error
	}{
		{
			name:   "successful cache hit",
			peerID: "peer123",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockLeaseRepository, mockCache *mocks.MockLeaseCache) {
				expectedLease := &models.Lease{
					TokenID:   12345,
					PeerID:    "peer123",
					ExpiresAt: time.Now().Add(time.Hour),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				mockCache.EXPECT().GetLeaseByPeerID(gomock.Any(), "peer123").Return(expectedLease, nil)
			},
			expectedLease: &models.Lease{
				TokenID:   12345,
				PeerID:    "peer123",
				ExpiresAt: time.Now().Add(time.Hour),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectedError: nil,
		},
		{
			name:   "cache miss, database hit",
			peerID: "peer456",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockLeaseRepository, mockCache *mocks.MockLeaseCache) {
				expectedLease := &models.Lease{
					TokenID:   67890,
					PeerID:    "peer456",
					ExpiresAt: time.Now().Add(time.Hour),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				mockCache.EXPECT().GetLeaseByPeerID(gomock.Any(), "peer456").Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetLeaseByPeerID(gomock.Any(), "peer456").Return(expectedLease, nil)
				mockCache.EXPECT().SetLease(gomock.Any(), expectedLease).Return(nil)
			},
			expectedLease: &models.Lease{
				TokenID:   67890,
				PeerID:    "peer456",
				ExpiresAt: time.Now().Add(time.Hour),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectedError: nil,
		},
		{
			name:   "cache miss, database miss",
			peerID: "peer789",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockLeaseRepository, mockCache *mocks.MockLeaseCache) {
				mockCache.EXPECT().GetLeaseByPeerID(gomock.Any(), "peer789").Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetLeaseByPeerID(gomock.Any(), "peer789").Return(nil, errors.New("not found"))
			},
			expectedLease: nil,
			expectedError: errors.New("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockLeaseRepository(ctrl)
			mockCache := mocks.NewMockLeaseCache(ctrl)
			logger := zap.NewNop()

			tt.mockSetup(ctrl, mockRepo, mockCache)

			hybridRepo := hybrid.NewLeaseRepository(mockRepo, mockCache, logger)

			result, err := hybridRepo.GetLeaseByPeerID(context.Background(), tt.peerID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLease.TokenID, result.TokenID)
				assert.Equal(t, tt.expectedLease.PeerID, result.PeerID)
			}
		})
	}
}

func TestLeaseRepository_GetLeaseByTokenID(t *testing.T) {
	tests := []struct {
		name          string
		tokenID       int64
		mockSetup     func(*gomock.Controller, *mocks.MockLeaseRepository, *mocks.MockLeaseCache)
		expectedLease *models.Lease
		expectedError error
	}{
		{
			name:    "successful cache hit",
			tokenID: 12345,
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockLeaseRepository, mockCache *mocks.MockLeaseCache) {
				expectedLease := &models.Lease{
					TokenID:   12345,
					PeerID:    "peer123",
					ExpiresAt: time.Now().Add(time.Hour),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				mockCache.EXPECT().GetLeaseByTokenID(gomock.Any(), int64(12345)).Return(expectedLease, nil)
			},
			expectedLease: &models.Lease{
				TokenID:   12345,
				PeerID:    "peer123",
				ExpiresAt: time.Now().Add(time.Hour),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectedError: nil,
		},
		{
			name:    "cache miss, database hit",
			tokenID: 67890,
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockLeaseRepository, mockCache *mocks.MockLeaseCache) {
				expectedLease := &models.Lease{
					TokenID:   67890,
					PeerID:    "peer456",
					ExpiresAt: time.Now().Add(time.Hour),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				mockCache.EXPECT().GetLeaseByTokenID(gomock.Any(), int64(67890)).Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetLeaseByTokenID(gomock.Any(), int64(67890)).Return(expectedLease, nil)
				mockCache.EXPECT().SetLease(gomock.Any(), expectedLease).Return(nil)
			},
			expectedLease: &models.Lease{
				TokenID:   67890,
				PeerID:    "peer456",
				ExpiresAt: time.Now().Add(time.Hour),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockLeaseRepository(ctrl)
			mockCache := mocks.NewMockLeaseCache(ctrl)
			logger := zap.NewNop()

			tt.mockSetup(ctrl, mockRepo, mockCache)

			hybridRepo := hybrid.NewLeaseRepository(mockRepo, mockCache, logger)

			result, err := hybridRepo.GetLeaseByTokenID(context.Background(), tt.tokenID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLease.TokenID, result.TokenID)
				assert.Equal(t, tt.expectedLease.PeerID, result.PeerID)
			}
		})
	}
}

func TestLeaseRepository_AllocateNewLease(t *testing.T) {
	tests := []struct {
		name          string
		peerID        string
		mockSetup     func(*gomock.Controller, *mocks.MockLeaseRepository, *mocks.MockLeaseCache)
		expectedLease *models.Lease
		expectedError error
	}{
		{
			name:   "successful allocation",
			peerID: "peer123",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockLeaseRepository, mockCache *mocks.MockLeaseCache) {
				expectedLease := &models.Lease{
					TokenID:   12345,
					PeerID:    "peer123",
					ExpiresAt: time.Now().Add(time.Hour),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				mockRepo.EXPECT().AllocateNewLease(gomock.Any(), "peer123").Return(expectedLease, nil)
				mockCache.EXPECT().SetLease(gomock.Any(), expectedLease).Return(nil)
			},
			expectedLease: &models.Lease{
				TokenID:   12345,
				PeerID:    "peer123",
				ExpiresAt: time.Now().Add(time.Hour),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectedError: nil,
		},
		{
			name:   "database error",
			peerID: "peer456",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockLeaseRepository, mockCache *mocks.MockLeaseCache) {
				mockRepo.EXPECT().AllocateNewLease(gomock.Any(), "peer456").Return(nil, errors.New("database error"))
			},
			expectedLease: nil,
			expectedError: errors.New("database error"),
		},
		{
			name:   "cache error after successful database operation",
			peerID: "peer789",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockLeaseRepository, mockCache *mocks.MockLeaseCache) {
				expectedLease := &models.Lease{
					TokenID:   67890,
					PeerID:    "peer789",
					ExpiresAt: time.Now().Add(time.Hour),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				mockRepo.EXPECT().AllocateNewLease(gomock.Any(), "peer789").Return(expectedLease, nil)
				mockCache.EXPECT().SetLease(gomock.Any(), expectedLease).Return(errors.New("cache error"))
			},
			expectedLease: &models.Lease{
				TokenID:   67890,
				PeerID:    "peer789",
				ExpiresAt: time.Now().Add(time.Hour),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectedError: nil, // Cache error should not fail the operation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockLeaseRepository(ctrl)
			mockCache := mocks.NewMockLeaseCache(ctrl)
			logger := zap.NewNop()

			tt.mockSetup(ctrl, mockRepo, mockCache)

			hybridRepo := hybrid.NewLeaseRepository(mockRepo, mockCache, logger)

			result, err := hybridRepo.AllocateNewLease(context.Background(), tt.peerID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLease.TokenID, result.TokenID)
				assert.Equal(t, tt.expectedLease.PeerID, result.PeerID)
			}
		})
	}
}

func TestLeaseRepository_RenewLease(t *testing.T) {
	tests := []struct {
		name          string
		tokenID       int64
		peerID        string
		mockSetup     func(*gomock.Controller, *mocks.MockLeaseRepository, *mocks.MockLeaseCache)
		expectedLease *models.Lease
		expectedError error
	}{
		{
			name:    "successful renewal",
			tokenID: 12345,
			peerID:  "peer123",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockLeaseRepository, mockCache *mocks.MockLeaseCache) {
				expectedLease := &models.Lease{
					TokenID:   12345,
					PeerID:    "peer123",
					ExpiresAt: time.Now().Add(time.Hour),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				mockRepo.EXPECT().RenewLease(gomock.Any(), int64(12345), "peer123").Return(expectedLease, nil)
				mockCache.EXPECT().SetLease(gomock.Any(), expectedLease).Return(nil)
			},
			expectedLease: &models.Lease{
				TokenID:   12345,
				PeerID:    "peer123",
				ExpiresAt: time.Now().Add(time.Hour),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectedError: nil,
		},
		{
			name:    "database error",
			tokenID: 67890,
			peerID:  "peer456",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockLeaseRepository, mockCache *mocks.MockLeaseCache) {
				mockRepo.EXPECT().RenewLease(gomock.Any(), int64(67890), "peer456").Return(nil, errors.New("database error"))
			},
			expectedLease: nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockLeaseRepository(ctrl)
			mockCache := mocks.NewMockLeaseCache(ctrl)
			logger := zap.NewNop()

			tt.mockSetup(ctrl, mockRepo, mockCache)

			hybridRepo := hybrid.NewLeaseRepository(mockRepo, mockCache, logger)

			result, err := hybridRepo.RenewLease(context.Background(), tt.tokenID, tt.peerID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLease.TokenID, result.TokenID)
				assert.Equal(t, tt.expectedLease.PeerID, result.PeerID)
			}
		})
	}
}

func TestLeaseRepository_ReleaseLease(t *testing.T) {
	tests := []struct {
		name          string
		tokenID       int64
		peerID        string
		mockSetup     func(*gomock.Controller, *mocks.MockLeaseRepository, *mocks.MockLeaseCache)
		expectedError error
	}{
		{
			name:    "successful release",
			tokenID: 12345,
			peerID:  "peer123",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockLeaseRepository, mockCache *mocks.MockLeaseCache) {
				mockRepo.EXPECT().ReleaseLease(gomock.Any(), int64(12345), "peer123").Return(nil)
				mockCache.EXPECT().DeleteLease(gomock.Any(), "peer123", int64(12345)).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:    "database error",
			tokenID: 67890,
			peerID:  "peer456",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockLeaseRepository, mockCache *mocks.MockLeaseCache) {
				mockRepo.EXPECT().ReleaseLease(gomock.Any(), int64(67890), "peer456").Return(errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
		{
			name:    "cache error after successful database operation",
			tokenID: 11111,
			peerID:  "peer789",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockLeaseRepository, mockCache *mocks.MockLeaseCache) {
				mockRepo.EXPECT().ReleaseLease(gomock.Any(), int64(11111), "peer789").Return(nil)
				mockCache.EXPECT().DeleteLease(gomock.Any(), "peer789", int64(11111)).Return(errors.New("cache error"))
			},
			expectedError: nil, // Cache error should not fail the operation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockLeaseRepository(ctrl)
			mockCache := mocks.NewMockLeaseCache(ctrl)
			logger := zap.NewNop()

			tt.mockSetup(ctrl, mockRepo, mockCache)

			hybridRepo := hybrid.NewLeaseRepository(mockRepo, mockCache, logger)

			err := hybridRepo.ReleaseLease(context.Background(), tt.tokenID, tt.peerID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
