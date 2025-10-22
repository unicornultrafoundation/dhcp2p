package ethereum

import (
	"context"
	"fmt"
	"sync"

	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/ports"
	"github.com/duchuongnguyen/dhcp2p/internal/app/infrastructure/config"
	initPkg "github.com/duchuongnguyen/dhcp2p/internal/app/infrastructure/init"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type AccountRepository struct {
	ks            *keystore.KeyStore
	password      string
	mu            sync.RWMutex
	cachedAccount *accounts.Account
	account       string
	logger        *zap.Logger
}

func NewAccountRepository(lc fx.Lifecycle, cfg *config.AppConfig, logger *zap.Logger) (*AccountRepository, error) {
	passwordPath := cfg.Password
	password, err := readPasswordFromFile(passwordPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read password: %w", err)
	}

	ks := keystore.NewKeyStore(cfg.Datadir, keystore.StandardScryptN, keystore.StandardScryptP)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Ensure keystore has at least one account
			if len(ks.Accounts()) == 0 {
				_, _, err := initPkg.CreateKeystoreAccount(ks, password, "Enter keystore password: ")
				if err != nil {
					return fmt.Errorf("failed to create keystore account: %w", err)
				}
			}
			return nil
		},
	})

	return &AccountRepository{ks: ks, password: password, account: cfg.Account}, nil
}

// GetCurrentAccount returns the account matching config.account.address, or the first account if not set or not found
func (s *AccountRepository) GetCurrentAccount() (accounts.Account, error) {
	s.mu.RLock()
	if s.cachedAccount != nil {
		account := *s.cachedAccount
		s.mu.RUnlock()
		return account, nil
	}
	s.mu.RUnlock()

	addressHex := s.account
	keystoreAccounts := s.ks.Accounts()

	if len(keystoreAccounts) == 0 {
		return accounts.Account{}, fmt.Errorf("no accounts found in keystore")
	}

	// If no specific address is configured, return the first account
	if addressHex == "" {
		s.mu.Lock()
		s.cachedAccount = &keystoreAccounts[0]
		account := keystoreAccounts[0]
		s.mu.Unlock()
		return account, nil
	}

	// Find account by address
	targetAddr := common.HexToAddress(addressHex)
	for _, acc := range keystoreAccounts {
		if acc.Address == targetAddr {
			s.mu.Lock()
			s.cachedAccount = &acc
			s.mu.Unlock()
			return acc, nil
		}
	}

	// If not found, fallback to first account with warning
	s.mu.Lock()
	s.cachedAccount = &keystoreAccounts[0]
	account := keystoreAccounts[0]
	s.mu.Unlock()
	s.logger.
		With(zap.String("using_address", addressHex)).
		With(zap.String("provided_address", addressHex)).
		Warn("Provided account not found, using first available account")

	return account, nil
}

// Implement AccountService interface
var _ ports.AccountService = (*AccountRepository)(nil)

func (s *AccountRepository) GetAddress() (string, error) {
	account, err := s.GetCurrentAccount()
	if err != nil {
		return "", fmt.Errorf("failed to get current account: %w", err)
	}
	return account.Address.Hex(), nil
}

// Sign the hash using ECDSA
func (s *AccountRepository) Sign(hash []byte) ([]byte, error) {
	if len(hash) == 0 {
		return nil, fmt.Errorf("hash cannot be empty")
	}

	s.mu.RLock()
	account := s.cachedAccount
	s.mu.RUnlock()

	if account == nil {
		var err error
		acc, err := s.GetCurrentAccount()
		if err != nil {
			return nil, fmt.Errorf("failed to get current account: %w", err)
		}
		account = &acc
	}

	signature, err := s.ks.SignHash(*account, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign hash: %w", err)
	}

	return signature, nil
}
