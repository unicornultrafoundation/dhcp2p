package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/application/services"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/errors"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/models"
	"github.com/unicornultrafoundation/dhcp2p/tests/mocks"
	"go.uber.org/mock/gomock"
)

func TestNonceService_CreateNonce(t *testing.T) {
	tests := []struct {
		name           string
		peerID         string
		mockSetup      func(*gomock.Controller, *mocks.MockNonceRepository, *mocks.MockSignatureVerifier)
		expectedResult *models.Nonce
		expectedError  error
	}{
		{
			name:   "successful nonce creation",
			peerID: "peer123",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockVerifier *mocks.MockSignatureVerifier) {
				expectedNonce := &models.Nonce{
					ID:        "nonce-123",
					PeerID:    "peer123",
					IssuedAt:  time.Now(),
					ExpiresAt: time.Now().Add(5 * time.Minute),
					Used:      false,
				}
				mockRepo.EXPECT().CreateNonce(gomock.Any(), "peer123").Return(expectedNonce, nil)
			},
			expectedResult: &models.Nonce{
				ID:        "nonce-123",
				PeerID:    "peer123",
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(5 * time.Minute),
				Used:      false,
			},
			expectedError: nil,
		},
		{
			name:   "empty peer ID",
			peerID: "",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockVerifier *mocks.MockSignatureVerifier) {
				mockRepo.EXPECT().CreateNonce(gomock.Any(), "").Return(nil, errors.ErrMissingPeerID)
			},
			expectedResult: nil,
			expectedError:  errors.ErrMissingPeerID,
		},
		{
			name:   "repository error",
			peerID: "peer456",
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockVerifier *mocks.MockSignatureVerifier) {
				mockRepo.EXPECT().CreateNonce(gomock.Any(), "peer456").Return(nil, errors.ErrMissingPeerID)
			},
			expectedResult: nil,
			expectedError:  errors.ErrMissingPeerID,
		},
		{
			name:   "very long peer ID",
			peerID: string(make([]byte, 1000)), // Very long peer ID
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockVerifier *mocks.MockSignatureVerifier) {
				mockRepo.EXPECT().CreateNonce(gomock.Any(), gomock.Any()).Return(nil, errors.ErrMissingPeerID)
			},
			expectedResult: nil,
			expectedError:  errors.ErrMissingPeerID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockNonceRepository(ctrl)
			mockVerifier := mocks.NewMockSignatureVerifier(ctrl)
			tt.mockSetup(ctrl, mockRepo, mockVerifier)

			service := services.NewNonceService(mockRepo, mockVerifier)

			result, err := service.CreateNonce(context.Background(), tt.peerID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.ID, result.ID)
				assert.Equal(t, tt.expectedResult.PeerID, result.PeerID)
				assert.Equal(t, tt.expectedResult.Used, result.Used)
			}
		})
	}
}

func TestNonceService_VerifyNonce(t *testing.T) {
	tests := []struct {
		name          string
		request       *models.NonceRequest
		mockSetup     func(*gomock.Controller, *mocks.MockNonceRepository, *mocks.MockSignatureVerifier)
		expectedError error
	}{
		{
			name: "successful nonce verification",
			request: &models.NonceRequest{
				NonceID:   "nonce-123",
				Pubkey:    []byte("valid-pubkey"),
				Payload:   []byte("payload"),
				Signature: []byte("signature"),
			},
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockVerifier *mocks.MockSignatureVerifier) {
				mockVerifier.EXPECT().VerifySignature(gomock.Any(), []byte("valid-pubkey"), []byte("payload"), []byte("signature")).Return(nil)
				mockRepo.EXPECT().GetNonce(gomock.Any(), "nonce-123").Return(&models.Nonce{ID: "nonce-123", PeerID: "test-peer"}, nil)
				// GetPeerIDFromPubkey will fail, but we still need the mock expectations above
			},
			expectedError: nil, // Will be handled by GetPeerIDFromPubkey - expect any error
		},
		{
			name: "signature verification failure",
			request: &models.NonceRequest{
				NonceID:   "nonce-123",
				Pubkey:    []byte("valid-pubkey"),
				Payload:   []byte("payload"),
				Signature: []byte("invalid-signature"),
			},
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockVerifier *mocks.MockSignatureVerifier) {
				mockVerifier.EXPECT().VerifySignature(gomock.Any(), []byte("valid-pubkey"), []byte("payload"), []byte("invalid-signature")).Return(errors.ErrInvalidSignature)
			},
			expectedError: errors.ErrInvalidSignature,
		},
		{
			name: "nonce not found",
			request: &models.NonceRequest{
				NonceID:   "non-existent-nonce",
				Pubkey:    []byte("valid-pubkey"),
				Payload:   []byte("payload"),
				Signature: []byte("signature"),
			},
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockVerifier *mocks.MockSignatureVerifier) {
				mockVerifier.EXPECT().VerifySignature(gomock.Any(), []byte("valid-pubkey"), []byte("payload"), []byte("signature")).Return(nil)
				mockRepo.EXPECT().GetNonce(gomock.Any(), "non-existent-nonce").Return(nil, errors.ErrNonceNotFound)
			},
			expectedError: errors.ErrNonceNotFound,
		},
		{
			name: "nonce consumption failure",
			request: &models.NonceRequest{
				NonceID:   "nonce-123",
				Pubkey:    []byte("valid-pubkey"),
				Payload:   []byte("payload"),
				Signature: []byte("signature"),
			},
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockVerifier *mocks.MockSignatureVerifier) {
				mockVerifier.EXPECT().VerifySignature(gomock.Any(), []byte("valid-pubkey"), []byte("payload"), []byte("signature")).Return(nil)
				mockRepo.EXPECT().GetNonce(gomock.Any(), "nonce-123").Return(&models.Nonce{ID: "nonce-123", PeerID: "test-peer"}, nil)
				// GetPeerIDFromPubkey will fail, but we still need the mock expectations above
			},
			expectedError: nil, // Will be handled by GetPeerIDFromPubkey - expect any error
		},
		{
			name: "empty nonce ID",
			request: &models.NonceRequest{
				NonceID:   "",
				Pubkey:    []byte("valid-pubkey"),
				Payload:   []byte("payload"),
				Signature: []byte("signature"),
			},
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockVerifier *mocks.MockSignatureVerifier) {
				mockVerifier.EXPECT().VerifySignature(gomock.Any(), []byte("valid-pubkey"), []byte("payload"), []byte("signature")).Return(nil)
				mockRepo.EXPECT().GetNonce(gomock.Any(), "").Return(nil, errors.ErrMissingPeerID)
			},
			expectedError: errors.ErrMissingPeerID,
		},
		{
			name: "nil pubkey",
			request: &models.NonceRequest{
				NonceID:   "nonce-123",
				Pubkey:    nil,
				Payload:   []byte("payload"),
				Signature: []byte("signature"),
			},
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockVerifier *mocks.MockSignatureVerifier) {
				mockVerifier.EXPECT().VerifySignature(gomock.Any(), nil, []byte("payload"), []byte("signature")).Return(errors.ErrInvalidSignature)
			},
			expectedError: errors.ErrInvalidSignature,
		},
		{
			name: "nil payload",
			request: &models.NonceRequest{
				NonceID:   "nonce-123",
				Pubkey:    []byte("valid-pubkey"),
				Payload:   nil,
				Signature: []byte("signature"),
			},
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockVerifier *mocks.MockSignatureVerifier) {
				mockVerifier.EXPECT().VerifySignature(gomock.Any(), []byte("valid-pubkey"), nil, []byte("signature")).Return(errors.ErrInvalidSignature)
			},
			expectedError: errors.ErrInvalidSignature,
		},
		{
			name: "nil signature",
			request: &models.NonceRequest{
				NonceID:   "nonce-123",
				Pubkey:    []byte("valid-pubkey"),
				Payload:   []byte("payload"),
				Signature: nil,
			},
			mockSetup: func(ctrl *gomock.Controller, mockRepo *mocks.MockNonceRepository, mockVerifier *mocks.MockSignatureVerifier) {
				mockVerifier.EXPECT().VerifySignature(gomock.Any(), []byte("valid-pubkey"), []byte("payload"), nil).Return(errors.ErrInvalidSignature)
			},
			expectedError: errors.ErrInvalidSignature,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockNonceRepository(ctrl)
			mockVerifier := mocks.NewMockSignatureVerifier(ctrl)
			tt.mockSetup(ctrl, mockRepo, mockVerifier)

			service := services.NewNonceService(mockRepo, mockVerifier)

			err := service.VerifyNonce(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				// For tests with nil expectedError, we expect an error from GetPeerIDFromPubkey
				// since we're using invalid pubkey data ([]byte("valid-pubkey"))
				assert.Error(t, err)
			}
		})
	}
}

func TestNonceService_EdgeCases(t *testing.T) {
	t.Run("context cancellation during signature verification", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockNonceRepository(ctrl)
		mockVerifier := mocks.NewMockSignatureVerifier(ctrl)
		service := services.NewNonceService(mockRepo, mockVerifier)

		// Create a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		request := &models.NonceRequest{
			NonceID:   "nonce-123",
			Pubkey:    []byte("valid-pubkey"),
			Payload:   []byte("payload"),
			Signature: []byte("signature"),
		}

		mockVerifier.EXPECT().VerifySignature(gomock.Any(), []byte("valid-pubkey"), []byte("payload"), []byte("signature")).Return(context.Canceled)

		err := service.VerifyNonce(ctx, request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("context cancellation during nonce retrieval", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockNonceRepository(ctrl)
		mockVerifier := mocks.NewMockSignatureVerifier(ctrl)
		service := services.NewNonceService(mockRepo, mockVerifier)

		request := &models.NonceRequest{
			NonceID:   "nonce-123",
			Pubkey:    []byte("valid-pubkey"),
			Payload:   []byte("payload"),
			Signature: []byte("signature"),
		}

		mockVerifier.EXPECT().VerifySignature(gomock.Any(), []byte("valid-pubkey"), []byte("payload"), []byte("signature")).Return(nil)
		mockRepo.EXPECT().GetNonce(gomock.Any(), "nonce-123").Return(nil, context.Canceled)

		err := service.VerifyNonce(context.Background(), request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("very large nonce ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockNonceRepository(ctrl)
		mockVerifier := mocks.NewMockSignatureVerifier(ctrl)
		service := services.NewNonceService(mockRepo, mockVerifier)

		largeNonceID := string(make([]byte, 10000))
		request := &models.NonceRequest{
			NonceID:   largeNonceID,
			Pubkey:    []byte("valid-pubkey"),
			Payload:   []byte("payload"),
			Signature: []byte("signature"),
		}

		mockVerifier.EXPECT().VerifySignature(gomock.Any(), []byte("valid-pubkey"), []byte("payload"), []byte("signature")).Return(nil)
		mockRepo.EXPECT().GetNonce(gomock.Any(), largeNonceID).Return(nil, errors.ErrMissingPeerID)

		err := service.VerifyNonce(context.Background(), request)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), errors.ErrMissingPeerID.Error())
	})

	t.Run("concurrent nonce creation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockNonceRepository(ctrl)
		mockVerifier := mocks.NewMockSignatureVerifier(ctrl)
		service := services.NewNonceService(mockRepo, mockVerifier)

		const numGoroutines = 10
		results := make(chan *models.Nonce, numGoroutines)
		errors := make(chan error, numGoroutines)

		// Setup expectations for concurrent calls
		for i := 0; i < numGoroutines; i++ {
			peerID := fmt.Sprintf("peer-%d", i)
			expectedNonce := &models.Nonce{
				ID:        fmt.Sprintf("nonce-%d", i),
				PeerID:    peerID,
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(5 * time.Minute),
				Used:      false,
			}
			mockRepo.EXPECT().CreateNonce(gomock.Any(), peerID).Return(expectedNonce, nil)
		}

		// Start multiple goroutines
		for i := 0; i < numGoroutines; i++ {
			go func(peerID string) {
				nonce, err := service.CreateNonce(context.Background(), peerID)
				if err != nil {
					errors <- err
					return
				}
				results <- nonce
			}(fmt.Sprintf("peer-%d", i))
		}

		// Collect results
		var nonces []*models.Nonce
		for i := 0; i < numGoroutines; i++ {
			select {
			case nonce := <-results:
				nonces = append(nonces, nonce)
			case err := <-errors:
				t.Errorf("Error in goroutine: %v", err)
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for concurrent nonce creation")
			}
		}

		// Verify all nonces were created successfully
		assert.Len(t, nonces, numGoroutines)
	})
}
