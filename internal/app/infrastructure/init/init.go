package init

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/duchuongnguyen/dhcp2p/internal/pkg/utils"
	"github.com/ethereum/go-ethereum/accounts/keystore"
)

func Init(repoPath string, out io.Writer) error {
	expPath, err := utils.ExpandHome(filepath.Clean(repoPath))
	if err != nil {
		return err
	}

	if err := os.MkdirAll(expPath, 0755); err != nil {
		return err
	}

	// Generate a new private key and keystore
	ksDir := filepath.Join(expPath, "keystore")
	if err := os.MkdirAll(ksDir, 0700); err != nil {
		return fmt.Errorf("failed to create keystore dir: %w", err)
	}
	ks := keystore.NewKeyStore(ksDir, keystore.StandardScryptN, keystore.StandardScryptP)
	account, _, _ := CreateKeystoreAccount(ks, "", "Enter keystore password: ")
	fmt.Fprintf(out, "Keystore created at %s\n", account.URL.Path)

	return nil
}
