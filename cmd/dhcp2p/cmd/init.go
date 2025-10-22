package cmd

import (
	"fmt"
	"os"

	"github.com/duchuongnguyen/dhcp2p/internal/app/infrastructure/flag"
	initPkg "github.com/duchuongnguyen/dhcp2p/internal/app/infrastructure/init"
	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the dhcp2p",
		Run: func(cmd *cobra.Command, args []string) {
			dataPath, _ := cmd.Flags().GetString(flag.DATADIR_FLAG)
			out := cmd.OutOrStdout()
			if err := initPkg.Init(dataPath, out); err != nil {
				fmt.Fprintf(out, "Error initializing: %v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}
