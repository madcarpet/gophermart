package config

import (
	"flag"

	"github.com/caarlos0/env"
)

// Application config structure
type Config struct {
	RunAddress            string `env:"RUN_ADDRESS"`
	DatabaseURI           string `env:"DATABASE_URI"`
	AccrualSystemAddress  string `env:"ACCURAL_SYSTEM_ADDRESS"`
	LogLevel              string `env:"LOG_LEVEL"`
	TokenKey              string `env:"TOKEN_KEY"`
	TokenTimeout          int    `env:"TOKEN_TIMEOUT"`
	OrdersQueueSize       int    `env:"ORDERS_QUEUE_SIZE"`
	AccrualWorkers        int    `env:"ACCURAL_WORKERS"`
	AccrualDelayedWorkers int    `env:"ACCURAL_DELAYED_WORKERS"`
	AccrualDelay          int    `env:"ACCURAL_DELAY"`
	AccrualDelayedBatch   int    `env:"ACCURAL_DELAYED_BATCH"`
	AccrualRequestRepeats int    `env:"ACCURAL_REQ_REPEATS"`
}

// Constructor for config structure, parses environment variable or cli arguments
func InitConfig() (*Config, error) {
	var config Config
	var cliConfig Config
	err := env.Parse(&config)
	if err != nil {
		return nil, err
	}

	flag.StringVar(&cliConfig.RunAddress, "a", "localhost:8080", "server IP address and TCP port (env:RUN_ADDRESS)")
	flag.StringVar(&cliConfig.DatabaseURI, "d", "postgresql://gopher:gopher@localhost:5432/gophermart", "database URI (env:DATABASE_URI)")
	flag.StringVar(&cliConfig.AccrualSystemAddress, "r", "http://localhost:8080/api/orders/", "accrual system IP address (env:ACCURAL_SYSTEM_ADDRESS)")
	flag.StringVar(&cliConfig.LogLevel, "l", "info", "logging level debug|info|warn|error (env:LOG_LEVEL)")
	flag.StringVar(&cliConfig.TokenKey, "k", "secretkey", "token secret key (env:TOKEN_KEY)")
	flag.IntVar(&cliConfig.TokenTimeout, "t", 3, "token timeout in hours (env:TOKEN_TIMEOUT)")
	flag.IntVar(&cliConfig.OrdersQueueSize, "q", 20, "accrual system client orders queue size (env:ORDERS_QUEUE_SIZE)")
	flag.IntVar(&cliConfig.AccrualWorkers, "aw", 3, "accrual system client workers number (env:ACCURAL_WORKERS)")
	flag.IntVar(&cliConfig.AccrualDelayedWorkers, "adw", 1, "accrual system client workers for delayed ordersnumber (env:ACCURAL_DELAYED_WORKERS)")
	flag.IntVar(&cliConfig.AccrualDelay, "dt", 2, "accrual system client delayed orders processing interval time in seconds (env:ACCURAL_DELAY)")
	flag.IntVar(&cliConfig.AccrualDelayedBatch, "dbs", 50, "accrual system client delayed orders processing batch size (env:ACCURAL_DELAYED_BATCH)")
	flag.IntVar(&cliConfig.AccrualRequestRepeats, "repeat", 3, "accrual system client request repeat times (env:ACCURAL_REQ_REPEATS)")
	flag.Parse()

	if config.RunAddress == "" {
		config.RunAddress = cliConfig.RunAddress
	}
	if config.DatabaseURI == "" {
		config.DatabaseURI = cliConfig.DatabaseURI
	}
	if config.AccrualSystemAddress == "" {
		config.AccrualSystemAddress = cliConfig.AccrualSystemAddress
	}
	if config.LogLevel == "" {
		config.LogLevel = cliConfig.LogLevel
	}
	if config.TokenKey == "" {
		config.TokenKey = cliConfig.TokenKey
	}
	if config.TokenTimeout == 0 {
		config.TokenTimeout = cliConfig.TokenTimeout
	}
	if config.OrdersQueueSize == 0 {
		config.OrdersQueueSize = cliConfig.OrdersQueueSize
	}
	if config.AccrualWorkers == 0 {
		config.AccrualWorkers = cliConfig.AccrualWorkers
	}
	if config.AccrualDelayedWorkers == 0 {
		config.AccrualDelayedWorkers = cliConfig.AccrualDelayedWorkers
	}
	if config.AccrualDelay == 0 {
		config.AccrualDelay = cliConfig.AccrualDelay
	}
	if config.AccrualDelayedBatch == 0 {
		config.AccrualDelayedBatch = cliConfig.AccrualDelayedBatch
	}
	if config.AccrualRequestRepeats == 0 {
		config.AccrualRequestRepeats = cliConfig.AccrualRequestRepeats
	}

	return &config, nil
}
