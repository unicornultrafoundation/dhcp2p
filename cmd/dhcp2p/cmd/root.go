package cmd

import (
	"github.com/duchuongnguyen/dhcp2p/internal/app/infrastructure/flag"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dhcp2p",
		Short: "DHCP2P",
		Long:  "DHCP2P - DHCP2P Service",
	}

	// Add flags
	cmd.PersistentFlags().StringP(flag.CONFIG_FLAG, flag.CONFIG_FLAG_SHORT, "", "Path to config file or directory")
	cmd.PersistentFlags().StringP(flag.DATADIR_FLAG, flag.DATA_DIR_FLAG_SHORT, "~/.dhcp2p", "Path to data directory")

	// Bind flags
	viper.BindPFlag("config", cmd.PersistentFlags().Lookup(flag.CONFIG_FLAG))
	viper.BindPFlag("datadir", cmd.PersistentFlags().Lookup(flag.DATADIR_FLAG))

	// Add commands
	cmd.AddCommand(initCmd())
	cmd.AddCommand(serveCmd())
	cmd.AddCommand(versionCmd())
	cmd.AddCommand(accountCmd())

	return cmd
}
