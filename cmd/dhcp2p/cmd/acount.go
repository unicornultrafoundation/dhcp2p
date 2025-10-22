package cmd

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/duchuongnguyen/dhcp2p/internal/app/infrastructure/flag"
	"github.com/duchuongnguyen/dhcp2p/internal/pkg/utils"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

var accountCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new account",
	Run: func(cmd *cobra.Command, args []string) {
		dataPath, _ := cmd.Flags().GetString(flag.DATADIR_FLAG)

		keystoreDir, err := getKeyStoreDir(dataPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		fmt.Print("Password: ")
		var password string
		fmt.Scanln(&password)
		ks := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
		account, err := ks.NewAccount(password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create account: %v\n", err)
			return
		}
		// Get the keystore file path
		keystorePath := filepath.Join(keystoreDir, account.URL.Path)
		fmt.Printf("New account created: %s\n", account.Address.Hex())
		fmt.Printf("Keystore file: %s\n", keystorePath)
	},
}

func getKeyStoreDir(dataPath string) (string, error) {
	expPath, err := utils.ExpandHome(filepath.Clean(dataPath))
	if err != nil {
		return "", err
	}

	keystoreDir := filepath.Join(expPath, "keystore")
	if _, err := os.Stat(keystoreDir); os.IsNotExist(err) {
		if err := os.MkdirAll(keystoreDir, 0700); err != nil {
			return "", fmt.Errorf("failed to create keystore directory: %w", err)
		}
	}

	return keystoreDir, nil
}

var accountListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all accounts in keystore",
	Run: func(cmd *cobra.Command, args []string) {
		dataPath, _ := cmd.Flags().GetString(flag.DATADIR_FLAG)

		keystoreDir, err := getKeyStoreDir(dataPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}

		ks := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
		accounts := ks.Accounts()
		if len(accounts) == 0 {
			fmt.Println("No accounts found.")
			return
		}
		fmt.Println("Accounts found:")
		fmt.Println("Address                                    | Keystore File")
		fmt.Println("------------------------------------------|----------------------------------")
		for _, acc := range accounts {
			address := acc.Address.Hex()
			keystorePath := filepath.Join(keystoreDir, acc.URL.Path)
			fmt.Printf("%-42s | %s\n", address, keystorePath)
		}
	},
}

var accountUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update password for an account",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dataPath, _ := cmd.Flags().GetString(flag.DATADIR_FLAG)

		address := args[0]
		keystoreDir, err := getKeyStoreDir(dataPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		ks := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
		for _, acc := range ks.Accounts() {
			if acc.Address.Hex() == address {
				fmt.Print("Old password: ")
				var oldPass string
				fmt.Scanln(&oldPass)
				fmt.Print("New password: ")
				var newPass string
				fmt.Scanln(&newPass)
				err := ks.Update(acc, oldPass, newPass)
				if err != nil {
					fmt.Printf("Update failed: %v\n", err)
					return
				}
				fmt.Println("Password updated.")
				return
			}
		}
		fmt.Println("Account not found.")
	},
}

var accountImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a keystore file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dataPath, _ := cmd.Flags().GetString(flag.DATADIR_FLAG)
		keystoreDir, err := getKeyStoreDir(dataPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		file := args[0]
		fmt.Print("Password: ")
		var password string
		fmt.Scanln(&password)
		ks := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
		keyjson, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Failed to read file: %v\n", err)
			return
		}
		_, err = ks.Import(keyjson, password, password)
		if err != nil {
			fmt.Printf("Import failed: %v\n", err)
			return
		}
		fmt.Println("Account imported.")
	},
}

var accountImportHexCmd = &cobra.Command{
	Use:   "import-hex",
	Short: "Import a private key in hex format",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dataPath, _ := cmd.Flags().GetString(flag.DATADIR_FLAG)
		keystoreDir, err := getKeyStoreDir(dataPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		hexkey := args[0]
		fmt.Print("Password: ")
		var password string
		fmt.Scanln(&password)
		privBytes, err := decodeHexKey(hexkey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid hex key: %v\n", err)
			return
		}
		privKey, err := toECDSA(privBytes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid private key: %v\n", err)
			return
		}
		ks := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
		account, err := ks.ImportECDSA(privKey, password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Import failed: %v\n", err)
			return
		}
		fmt.Printf("Account imported: %s\n", account.Address.Hex())
	},
}

func decodeHexKey(hexkey string) ([]byte, error) {
	if len(hexkey) >= 2 && hexkey[:2] == "0x" {
		hexkey = hexkey[2:]
	}
	return hex.DecodeString(hexkey)
}

func toECDSA(privBytes []byte) (*ecdsa.PrivateKey, error) {
	return ethcrypto.ToECDSA(privBytes)
}

func accountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Account management commands",
		Long:  "Commands to manage Subnet accounts and transactions.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(accountListCmd)
	cmd.AddCommand(accountUpdateCmd)
	cmd.AddCommand(accountImportCmd)
	cmd.AddCommand(accountImportHexCmd)
	cmd.AddCommand(accountCreateCmd)

	return cmd
}
