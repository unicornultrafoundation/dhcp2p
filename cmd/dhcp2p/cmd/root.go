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

	// Bind flags
	viper.BindPFlag("config", cmd.PersistentFlags().Lookup(flag.CONFIG_FLAG))

	// Add commands
	cmd.AddCommand(serveCmd())
	cmd.AddCommand(versionCmd())

	return cmd
}
