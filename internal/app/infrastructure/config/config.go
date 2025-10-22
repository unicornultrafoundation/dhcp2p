package config

import (
	"fmt"

	"github.com/duchuongnguyen/dhcp2p/internal/app/infrastructure/flag"

	"github.com/spf13/viper"
)

const (
	ENV_PREFIX = "DHCP2P"
)

type AppConfig struct {
	Port                 int    `mapstructure:"port"`
	LogLevel             string `mapstructure:"log_level"`
	Datadir              string `mapstructure:"datadir"`
	Password             string `mapstructure:"password"`
	Account              string `mapstructure:"account"`
	DatabaseURL          string `mapstructure:"database_url"`
	RedisURL             string `mapstructure:"redis_url"`
	NonceTTL             int    `mapstructure:"nonce_ttl"`              // in minutes
	NonceCleanerInterval int    `mapstructure:"nonce_cleaner_interval"` // in minutes
	LeaseTTL             int    `mapstructure:"lease_ttl"`              // in minutes
	MaxLeaseRetries      int    `mapstructure:"max_lease_retries"`
	LeaseRetryDelay      int    `mapstructure:"lease_retry_delay"` // in milliseconds

	// Redis Configuration
	RedisMaxRetries   int `mapstructure:"redis_max_retries"`
	RedisPoolSize     int `mapstructure:"redis_pool_size"`
	RedisMinIdleConns int `mapstructure:"redis_min_idle_conns"`
	RedisDialTimeout  int `mapstructure:"redis_dial_timeout"`  // seconds
	RedisReadTimeout  int `mapstructure:"redis_read_timeout"`  // seconds
	RedisWriteTimeout int `mapstructure:"redis_write_timeout"` // seconds

	// Cache Configuration
	CacheEnabled    bool `mapstructure:"cache_enabled"`
	CacheDefaultTTL int  `mapstructure:"cache_default_ttl"` // minutes
}

func NewAppConfig() (*AppConfig, error) {
	v := viper.GetViper()

	// Set environment prefix
	v.SetEnvPrefix(ENV_PREFIX)
	v.AutomaticEnv()

	// Load config file if exists
	configPath := v.GetString(flag.CONFIG_FLAG)
	if configPath != "" {
		// If user passed a file
		v.SetConfigFile(configPath)
	} else {
		// Fallback: look for config.yaml in current dir
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath("/etc/dhcp2p/") // optional global config path
	}

	// Try to read file (ignore if not found)
	if err := v.ReadInConfig(); err != nil {
		// Ignore "not found", but fail on parsing error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var c AppConfig
	if err := v.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &c, nil
}
