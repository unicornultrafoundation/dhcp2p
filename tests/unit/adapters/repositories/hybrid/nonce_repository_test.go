package hybrid

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/adapters/repositories/hybrid"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/models"
	"github.com/unicornultrafoundation/dhcp2p/tests/mocks"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
)

func TestNonceRepository_GetNonce(t *testing.T) {
	tests := []struct {
		name          string
		nonceID       string
		mockSetup     func(*gomock.Controller, *mocks.MockNonceRepository, *mocks.MockNonceCache)
		expectedNonce *models.Nonce
		expectedError error
	}{
		{
			name:    "successful cache hit",
			nonceID: "test-nonce-1",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockCache *mocks.MockNonceCache) {
				expectedNonce := &models.Nonce{
					ID:        "test-nonce-1",
					PeerID:    "peer123",
					IssuedAt:  time.Now(),
					ExpiresAt: time.Now().Add(time.Hour),
					Used:      false,
				}
				mockCache.EXPECT().GetNonce(gomock.Any(), "test-nonce-1").Return(expectedNonce, nil)
			},
			expectedNonce: &models.Nonce{
				ID:        "test-nonce-1",
				PeerID:    "peer123",
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      false,
			},
			expectedError: nil,
		},
		{
			name:    "cache miss, database hit",
			nonceID: "test-nonce-2",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockCache *mocks.MockNonceCache) {
				expectedNonce := &models.Nonce{
					ID:        "test-nonce-2",
					PeerID:    "peer456",
					IssuedAt:  time.Now(),
					ExpiresAt: time.Now().Add(time.Hour),
					Used:      false,
				}
				mockCache.EXPECT().GetNonce(gomock.Any(), "test-nonce-2").Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetNonce(gomock.Any(), "test-nonce-2").Return(expectedNonce, nil)
				mockCache.EXPECT().CreateNonce(gomock.Any(), expectedNonce).Return(nil)
			},
			expectedNonce: &models.Nonce{
				ID:        "test-nonce-2",
				PeerID:    "peer456",
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      false,
			},
			expectedError: nil,
		},
		{
			name:    "cache miss, database miss",
			nonceID: "test-nonce-3",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockCache *mocks.MockNonceCache) {
				mockCache.EXPECT().GetNonce(gomock.Any(), "test-nonce-3").Return(nil, errors.New("not found"))
				mockRepo.EXPECT().GetNonce(gomock.Any(), "test-nonce-3").Return(nil, errors.New("not found"))
			},
			expectedNonce: nil,
			expectedError: errors.New("not found"),
		},
		{
			name:    "cache error, database success",
			nonceID: "test-nonce-4",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockCache *mocks.MockNonceCache) {
				expectedNonce := &models.Nonce{
					ID:        "test-nonce-4",
					PeerID:    "peer789",
					IssuedAt:  time.Now(),
					ExpiresAt: time.Now().Add(time.Hour),
					Used:      false,
				}
				mockCache.EXPECT().GetNonce(gomock.Any(), "test-nonce-4").Return(nil, errors.New("cache error"))
				mockRepo.EXPECT().GetNonce(gomock.Any(), "test-nonce-4").Return(expectedNonce, nil)
				mockCache.EXPECT().CreateNonce(gomock.Any(), expectedNonce).Return(nil)
			},
			expectedNonce: &models.Nonce{
				ID:        "test-nonce-4",
				PeerID:    "peer789",
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      false,
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockNonceRepository(ctrl)
			mockCache := mocks.NewMockNonceCache(ctrl)
			logger := zap.NewNop()

			tt.mockSetup(ctrl, mockRepo, mockCache)

			hybridRepo := hybrid.NewNonceRepository(mockRepo, mockCache, logger)

			result, err := hybridRepo.GetNonce(context.Background(), tt.nonceID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedNonce.ID, result.ID)
				assert.Equal(t, tt.expectedNonce.PeerID, result.PeerID)
				assert.Equal(t, tt.expectedNonce.Used, result.Used)
			}
		})
	}
}

func TestNonceRepository_CreateNonce(t *testing.T) {
	tests := []struct {
		name          string
		peerID        string
		mockSetup     func(*gomock.Controller, *mocks.MockNonceRepository, *mocks.MockNonceCache)
		expectedNonce *models.Nonce
		expectedError error
	}{
		{
			name:   "successful creation",
			peerID: "peer123",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockCache *mocks.MockNonceCache) {
				expectedNonce := &models.Nonce{
					ID:        "new-nonce-1",
					PeerID:    "peer123",
					IssuedAt:  time.Now(),
					ExpiresAt: time.Now().Add(time.Hour),
					Used:      false,
				}
				mockRepo.EXPECT().CreateNonce(gomock.Any(), "peer123").Return(expectedNonce, nil)
				mockCache.EXPECT().CreateNonce(gomock.Any(), expectedNonce).Return(nil)
			},
			expectedNonce: &models.Nonce{
				ID:        "new-nonce-1",
				PeerID:    "peer123",
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      false,
			},
			expectedError: nil,
		},
		{
			name:   "database error",
			peerID: "peer456",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockCache *mocks.MockNonceCache) {
				mockRepo.EXPECT().CreateNonce(gomock.Any(), "peer456").Return(nil, errors.New("database error"))
			},
			expectedNonce: nil,
			expectedError: errors.New("database error"),
		},
		{
			name:   "cache error after successful database creation",
			peerID: "peer789",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockCache *mocks.MockNonceCache) {
				expectedNonce := &models.Nonce{
					ID:        "new-nonce-2",
					PeerID:    "peer789",
					IssuedAt:  time.Now(),
					ExpiresAt: time.Now().Add(time.Hour),
					Used:      false,
				}
				mockRepo.EXPECT().CreateNonce(gomock.Any(), "peer789").Return(expectedNonce, nil)
				mockCache.EXPECT().CreateNonce(gomock.Any(), expectedNonce).Return(errors.New("cache error"))
			},
			expectedNonce: &models.Nonce{
				ID:        "new-nonce-2",
				PeerID:    "peer789",
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      false,
			},
			expectedError: nil, // Cache error should not fail the operation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockNonceRepository(ctrl)
			mockCache := mocks.NewMockNonceCache(ctrl)
			logger := zap.NewNop()

			tt.mockSetup(ctrl, mockRepo, mockCache)

			hybridRepo := hybrid.NewNonceRepository(mockRepo, mockCache, logger)

			result, err := hybridRepo.CreateNonce(context.Background(), tt.peerID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedNonce.ID, result.ID)
				assert.Equal(t, tt.expectedNonce.PeerID, result.PeerID)
				assert.Equal(t, tt.expectedNonce.Used, result.Used)
			}
		})
	}
}

func TestNonceRepository_ConsumeNonce(t *testing.T) {
	tests := []struct {
		name          string
		nonceID       string
		peerID        string
		mockSetup     func(*gomock.Controller, *mocks.MockNonceRepository, *mocks.MockNonceCache)
		expectedError error
	}{
		{
			name:    "successful consumption",
			nonceID: "nonce-1",
			peerID:  "peer123",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockCache *mocks.MockNonceCache) {
				mockRepo.EXPECT().ConsumeNonce(gomock.Any(), "nonce-1", "peer123").Return(nil)
				mockCache.EXPECT().DeleteNonce(gomock.Any(), "nonce-1").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:    "database error",
			nonceID: "nonce-2",
			peerID:  "peer456",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockCache *mocks.MockNonceCache) {
				mockRepo.EXPECT().ConsumeNonce(gomock.Any(), "nonce-2", "peer456").Return(errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
		{
			name:    "cache error after successful database operation",
			nonceID: "nonce-3",
			peerID:  "peer789",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockCache *mocks.MockNonceCache) {
				mockRepo.EXPECT().ConsumeNonce(gomock.Any(), "nonce-3", "peer789").Return(nil)
				mockCache.EXPECT().DeleteNonce(gomock.Any(), "nonce-3").Return(errors.New("cache error"))
			},
			expectedError: nil, // Cache error should not fail the operation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockNonceRepository(ctrl)
			mockCache := mocks.NewMockNonceCache(ctrl)
			logger := zap.NewNop()

			tt.mockSetup(ctrl, mockRepo, mockCache)

			hybridRepo := hybrid.NewNonceRepository(mockRepo, mockCache, logger)

			err := hybridRepo.ConsumeNonce(context.Background(), tt.nonceID, tt.peerID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNonceRepository_DeleteExpiredNonces(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*gomock.Controller, *mocks.MockNonceRepository, *mocks.MockNonceCache)
		expectedError error
	}{
		{
			name: "successful cleanup",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockCache *mocks.MockNonceCache) {
				mockRepo.EXPECT().DeleteExpiredNonces(gomock.Any()).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "database error",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockCache *mocks.MockNonceCache) {
				mockRepo.EXPECT().DeleteExpiredNonces(gomock.Any()).Return(errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockNonceRepository(ctrl)
			mockCache := mocks.NewMockNonceCache(ctrl)
			logger := zap.NewNop()

			tt.mockSetup(ctrl, mockRepo, mockCache)

			hybridRepo := hybrid.NewNonceRepository(mockRepo, mockCache, logger)

			err := hybridRepo.DeleteExpiredNonces(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
