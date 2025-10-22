package init

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
)

// CreateKeystoreAccount prompts for password if needed and creates a new keystore account
func CreateKeystoreAccount(ks *keystore.KeyStore, password string, promptMsg string) (accounts.Account, string, error) {
	var err error
	if password == "" {
		password, err = PromptPassword(promptMsg)
		if err != nil {
			return accounts.Account{}, "", err
		}
	}
	acc, err := ks.NewAccount(password)
	if err != nil {
		return accounts.Account{}, "", fmt.Errorf("failed to create new keystore account: %w", err)
	}
	return acc, password, nil
}

// PromptPassword prompts the user for a password with the given message.
func PromptPassword(message string) (string, error) {
	fmt.Print(message)
	reader := bufio.NewReader(os.Stdin)
	pw, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	return strings.TrimSpace(pw), nil
}
