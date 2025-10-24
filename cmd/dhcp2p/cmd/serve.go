package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/unicornultrafoundation/dhcp2p/internal/app"
	"github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure/flag"
)

func serveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve the dhcp2p",
		Run: func(cmd *cobra.Command, args []string) {
			application := app.NewApp()
			application.Run()
		},
	}

	// Add flags
	cmd.Flags().IntP(flag.PORT_FLAG, flag.PORT_FLAG_SHORT, 0, "Port to run the server on")
	cmd.Flags().StringP(flag.LOG_LEVEL_FLAG, flag.LOG_LEVEL_FLAG_SHORT, "", "Log level")
	cmd.Flags().StringP(flag.DATABASE_URL_FLAG, flag.DATABASE_URL_FLAG_SHORT, "", "Database URL")
	cmd.Flags().StringP(flag.REDIS_URL_FLAG, flag.REDIS_URL_FLAG_SHORT, "", "Redis URL")
	cmd.Flags().StringP(flag.REDIS_PASSWORD_FLAG, flag.REDIS_PASSWORD_FLAG_SHORT, "", "Redis Password")
	cmd.Flags().IntP(flag.NONCE_TTL_FLAG, flag.NONCE_TTL_FLAG_SHORT, 0, "Nonce TTL")                                        // in minutes
	cmd.Flags().IntP(flag.NONCE_CLEANER_INTERVAL_FLAG, flag.NONCE_CLEANER_INTERVAL_FLAG_SHORT, 0, "Nonce Cleaner Interval") // in minutes
	cmd.Flags().IntP(flag.LEASE_TTL_FLAG, flag.LEASE_TTL_FLAG_SHORT, 0, "Lease TTL")                                        // in minutes
	cmd.Flags().IntP(flag.MAX_LEASE_RETRIES_FLAG, flag.MAX_LEASE_RETRIES_FLAG_SHORT, 0, "Max Lease Retries")
	cmd.Flags().IntP(flag.LEASE_RETRY_DELAY_FLAG, flag.LEASE_RETRY_DELAY_FLAG_SHORT, 0, "Lease Retry Delay") // in milliseconds

	// Required flags
	cmd.MarkFlagRequired(flag.DATABASE_URL_FLAG)
	cmd.MarkFlagRequired(flag.REDIS_URL_FLAG)

	// Bind flags
	viper.BindPFlag("port", cmd.Flags().Lookup(flag.PORT_FLAG))
	viper.BindPFlag("log", cmd.Flags().Lookup(flag.LOG_LEVEL_FLAG))
	viper.BindPFlag("database_url", cmd.Flags().Lookup(flag.DATABASE_URL_FLAG))
	viper.BindPFlag("redis_url", cmd.Flags().Lookup(flag.REDIS_URL_FLAG))
	viper.BindPFlag("redis_password", cmd.Flags().Lookup(flag.REDIS_PASSWORD_FLAG))
	viper.BindPFlag("nonce_ttl", cmd.Flags().Lookup(flag.NONCE_TTL_FLAG))
	viper.BindPFlag("nonce_cleaner_interval", cmd.Flags().Lookup(flag.NONCE_CLEANER_INTERVAL_FLAG))
	viper.BindPFlag("lease_ttl", cmd.Flags().Lookup(flag.LEASE_TTL_FLAG))
	viper.BindPFlag("max_lease_retries", cmd.Flags().Lookup(flag.MAX_LEASE_RETRIES_FLAG))
	viper.BindPFlag("lease_retry_delay", cmd.Flags().Lookup(flag.LEASE_RETRY_DELAY_FLAG))

	return cmd
}
