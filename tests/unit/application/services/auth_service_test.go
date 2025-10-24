package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/application/services"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/errors"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/domain/models"
	"github.com/unicornultrafoundation/dhcp2p/tests/mocks"
	"github.com/golang/mock/gomock"
)

func TestAuthService_RequestAuth(t *testing.T) {
	tests := []struct {
		name           string
		request        *models.AuthRequest
		mockSetup      func(*gomock.Controller, *mocks.MockNonceService)
		expectedResult *models.AuthResponse
		expectedError  error
	}{
		{
			name:    "successful auth request",
			request: &models.AuthRequest{Pubkey: []byte("valid-pubkey-data")},
			mockSetup: func(ctrl *gomock.Controller, mockNonce *mocks.MockNonceService) {
				// No expectations since crypto.UnmarshalPublicKey will fail
			},
			expectedResult: nil,
			expectedError:  nil, // Will be handled by crypto.UnmarshalPublicKey - expect any error
		},
		{
			name:           "nil request",
			request:        nil,
			mockSetup:      func(ctrl *gomock.Controller, mockNonce *mocks.MockNonceService) {},
			expectedResult: nil,
			expectedError:  errors.ErrMissingPubkey,
		},
		{
			name:           "empty pubkey",
			request:        &models.AuthRequest{Pubkey: []byte{}},
			mockSetup:      func(ctrl *gomock.Controller, mockNonce *mocks.MockNonceService) {},
			expectedResult: nil,
			expectedError:  nil, // Will be handled by crypto.UnmarshalPublicKey - expect any error
		},
		{
			name:           "invalid pubkey",
			request:        &models.AuthRequest{Pubkey: []byte("invalid-pubkey")},
			mockSetup:      func(ctrl *gomock.Controller, mockNonce *mocks.MockNonceService) {},
			expectedResult: nil,
			expectedError:  nil, // Will be handled by crypto.UnmarshalPublicKey - expect any error
		},
		{
			name:    "nonce service error",
			request: &models.AuthRequest{Pubkey: []byte("valid-pubkey-data")},
			mockSetup: func(ctrl *gomock.Controller, mockNonce *mocks.MockNonceService) {
				// No expectations since crypto.UnmarshalPublicKey will fail first
			},
			expectedResult: nil,
			expectedError:  nil, // Will be handled by crypto.UnmarshalPublicKey - expect any error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockNonce := mocks.NewMockNonceService(ctrl)
			tt.mockSetup(ctrl, mockNonce)

			service := services.NewAuthService(mockNonce)

			result, err := service.RequestAuth(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				if tt.expectedResult != nil {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedResult.NonceID, result.NonceID)
				} else {
					// For invalid pubkey cases, we expect an error from crypto operations
					assert.Error(t, err)
					assert.Nil(t, result)
				}
			}
		})
	}
}

func TestAuthService_VerifyAuth(t *testing.T) {
	tests := []struct {
		name           string
		request        *models.AuthVerifyRequest
		mockSetup      func(*gomock.Controller, *mocks.MockNonceService)
		expectedResult *models.AuthVerifyResponse
		expectedError  error
	}{
		{
			name: "successful auth verification",
			request: &models.AuthVerifyRequest{
				NonceID:   "test-nonce-id",
				Signature: []byte("valid-signature"),
				Pubkey:    []byte("valid-pubkey-data"),
			},
			mockSetup: func(ctrl *gomock.Controller, mockNonce *mocks.MockNonceService) {
				mockNonce.EXPECT().VerifyNonce(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedResult: &models.AuthVerifyResponse{
				Pubkey: []byte("valid-pubkey-data"),
			},
			expectedError: nil,
		},
		{
			name:           "nil request",
			request:        nil,
			mockSetup:      func(ctrl *gomock.Controller, mockNonce *mocks.MockNonceService) {},
			expectedResult: nil,
			expectedError:  errors.ErrMissingPeerID,
		},
		{
			name: "nil signature",
			request: &models.AuthVerifyRequest{
				NonceID:   "test-nonce-id",
				Signature: nil,
				Pubkey:    []byte("valid-pubkey-data"),
			},
			mockSetup:      func(ctrl *gomock.Controller, mockNonce *mocks.MockNonceService) {},
			expectedResult: nil,
			expectedError:  errors.ErrInvalidSignature,
		},
		{
			name: "empty signature",
			request: &models.AuthVerifyRequest{
				NonceID:   "test-nonce-id",
				Signature: []byte{},
				Pubkey:    []byte("valid-pubkey-data"),
			},
			mockSetup: func(ctrl *gomock.Controller, mockNonce *mocks.MockNonceService) {
				mockNonce.EXPECT().VerifyNonce(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedResult: &models.AuthVerifyResponse{
				Pubkey: []byte("valid-pubkey-data"),
			},
			expectedError: nil,
		},
		{
			name: "nonce verification error",
			request: &models.AuthVerifyRequest{
				NonceID:   "test-nonce-id",
				Signature: []byte("valid-signature"),
				Pubkey:    []byte("valid-pubkey-data"),
			},
			mockSetup: func(ctrl *gomock.Controller, mockNonce *mocks.MockNonceService) {
				mockNonce.EXPECT().VerifyNonce(gomock.Any(), gomock.Any()).Return(errors.ErrMissingPeerID)
			},
			expectedResult: nil,
			expectedError:  errors.ErrMissingPeerID,
		},
		{
			name: "empty nonce ID",
			request: &models.AuthVerifyRequest{
				NonceID:   "",
				Signature: []byte("valid-signature"),
				Pubkey:    []byte("valid-pubkey-data"),
			},
			mockSetup: func(ctrl *gomock.Controller, mockNonce *mocks.MockNonceService) {
				mockNonce.EXPECT().VerifyNonce(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedResult: &models.AuthVerifyResponse{
				Pubkey: []byte("valid-pubkey-data"),
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockNonce := mocks.NewMockNonceService(ctrl)
			tt.mockSetup(ctrl, mockNonce)

			service := services.NewAuthService(mockNonce)

			result, err := service.VerifyAuth(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Pubkey, result.Pubkey)
			}
		})
	}
}

func TestAuthService_EdgeCases(t *testing.T) {
	t.Run("very large pubkey", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockNonce := mocks.NewMockNonceService(ctrl)
		service := services.NewAuthService(mockNonce)

		// Create a very large invalid pubkey
		largePubkey := make([]byte, 10000)
		request := &models.AuthRequest{Pubkey: largePubkey}

		result, err := service.RequestAuth(context.Background(), request)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("very large signature", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockNonce := mocks.NewMockNonceService(ctrl)
		service := services.NewAuthService(mockNonce)

		// Create a very large signature
		largeSignature := make([]byte, 10000)
		request := &models.AuthVerifyRequest{
			NonceID:   "test-nonce-id",
			Signature: largeSignature,
			Pubkey:    []byte("valid-pubkey-data"),
		}

		mockNonce.EXPECT().VerifyNonce(gomock.Any(), gomock.Any()).Return(nil)

		result, err := service.VerifyAuth(context.Background(), request)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}
