package config

import (
	"fmt"

	"github.com/unicornultrafoundation/dhcp2p/internal/app/infrastructure/flag"

	"github.com/spf13/viper"
)

const (
	ENV_PREFIX = "DHCP2P"
)

type AppConfig struct {
	Port                 int    `mapstructure:"port"`
	LogLevel             string `mapstructure:"log_level"`
	DatabaseURL          string `mapstructure:"database_url"`
	RedisURL             string `mapstructure:"redis_url"`
	RedisPassword        string `mapstructure:"redis_password"`
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

	// PostgreSQL Pool Configuration
	DBMaxConns          int `mapstructure:"db_max_conns"`           // maximum number of connections in the pool
	DBMinConns          int `mapstructure:"db_min_conns"`           // minimum number of connections in the pool
	DBMaxConnLifetime   int `mapstructure:"db_max_conn_lifetime"`   // maximum lifetime of a connection in minutes
	DBMaxConnIdleTime   int `mapstructure:"db_max_conn_idle_time"`  // maximum idle time of a connection in minutes
	DBHealthCheckPeriod int `mapstructure:"db_health_check_period"` // health check period in seconds

	// Rate Limiting Configuration
	RateLimitEnabled           bool     `mapstructure:"rate_limit_enabled"`             // enable/disable rate limiting
	RateLimitRequestsPerMinute int      `mapstructure:"rate_limit_requests_per_minute"` // requests per minute per IP
	RateLimitBurst             int      `mapstructure:"rate_limit_burst"`               // burst capacity for token bucket
	RateLimitTrustedProxies    []string `mapstructure:"rate_limit_trusted_proxies"`     // trusted proxy IPs for header validation
}

// NewDefaultAppConfig returns an AppConfig with all default values
func NewDefaultAppConfig() *AppConfig {
	return &AppConfig{
		// Server Configuration
		Port:     8088,
		LogLevel: "info",

		// Nonce Configuration
		NonceTTL:             5, // minutes
		NonceCleanerInterval: 5, // minutes

		// Lease Configuration
		LeaseTTL:        120, // minutes
		MaxLeaseRetries: 3,
		LeaseRetryDelay: 500, // milliseconds

		// Redis Configuration
		RedisMaxRetries:   3,
		RedisPoolSize:     10,
		RedisMinIdleConns: 5,
		RedisDialTimeout:  5, // seconds
		RedisReadTimeout:  3, // seconds
		RedisWriteTimeout: 3, // seconds

		// Cache Configuration
		CacheEnabled:    true,
		CacheDefaultTTL: 30, // minutes

		// PostgreSQL Pool Configuration
		DBMaxConns:          25,
		DBMinConns:          5,
		DBMaxConnLifetime:   30, // minutes
		DBMaxConnIdleTime:   5,  // minutes
		DBHealthCheckPeriod: 30, // seconds

		// Rate Limiting Configuration
		RateLimitEnabled:           true,
		RateLimitRequestsPerMinute: 100,
		RateLimitBurst:             20,
		RateLimitTrustedProxies:    []string{},
	}
}

func NewAppConfig() (*AppConfig, error) {
	v := viper.GetViper()

	// Set environment prefix
	v.SetEnvPrefix(ENV_PREFIX)
	v.AutomaticEnv()

	// Set default values from our centralized defaults
	defaults := NewDefaultAppConfig()
	v.SetDefault("port", defaults.Port)
	v.SetDefault("log_level", defaults.LogLevel)
	v.SetDefault("nonce_ttl", defaults.NonceTTL)
	v.SetDefault("nonce_cleaner_interval", defaults.NonceCleanerInterval)
	v.SetDefault("lease_ttl", defaults.LeaseTTL)
	v.SetDefault("max_lease_retries", defaults.MaxLeaseRetries)
	v.SetDefault("lease_retry_delay", defaults.LeaseRetryDelay)
	v.SetDefault("redis_max_retries", defaults.RedisMaxRetries)
	v.SetDefault("redis_pool_size", defaults.RedisPoolSize)
	v.SetDefault("redis_min_idle_conns", defaults.RedisMinIdleConns)
	v.SetDefault("redis_dial_timeout", defaults.RedisDialTimeout)
	v.SetDefault("redis_read_timeout", defaults.RedisReadTimeout)
	v.SetDefault("redis_write_timeout", defaults.RedisWriteTimeout)
	v.SetDefault("cache_enabled", defaults.CacheEnabled)
	v.SetDefault("cache_default_ttl", defaults.CacheDefaultTTL)
	v.SetDefault("db_max_conns", defaults.DBMaxConns)
	v.SetDefault("db_min_conns", defaults.DBMinConns)
	v.SetDefault("db_max_conn_lifetime", defaults.DBMaxConnLifetime)
	v.SetDefault("db_max_conn_idle_time", defaults.DBMaxConnIdleTime)
	v.SetDefault("db_health_check_period", defaults.DBHealthCheckPeriod)
	v.SetDefault("rate_limit_enabled", defaults.RateLimitEnabled)
	v.SetDefault("rate_limit_requests_per_minute", defaults.RateLimitRequestsPerMinute)
	v.SetDefault("rate_limit_burst", defaults.RateLimitBurst)
	v.SetDefault("rate_limit_trusted_proxies", defaults.RateLimitTrustedProxies)

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
