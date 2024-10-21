package app

import (
	"context"
	"time"

	"github.com/madcarpet/gophermart/internal/accrual"
	"github.com/madcarpet/gophermart/internal/authorization/jwt"
	"github.com/madcarpet/gophermart/internal/config"
	"github.com/madcarpet/gophermart/internal/handlers"
	"github.com/madcarpet/gophermart/internal/logger"
	"github.com/madcarpet/gophermart/internal/models"
	"github.com/madcarpet/gophermart/internal/storage"
	"github.com/madcarpet/gophermart/internal/storage/postgresql"
	"go.uber.org/zap"
)

type App struct {
	config  config.Config
	storage storage.Storage
}

// NewApp creates a new App instance with the given config
func NewApp(cfg config.Config) *App {
	return &App{config: cfg}
}

// Start App
func (a *App) Start(ctx context.Context) error {
	logger.LoggerInit(a.config.LogLevel)
	logger.Log.Info("Starting application",
		zap.String("run_address", a.config.RunAddress),
		zap.String("database_uri", a.config.DatabaseUri),
		zap.String("accural_system_address", a.config.AccrualSystemAddress),
		zap.String("log_level", a.config.LogLevel),
		zap.String("token_key", a.config.TokenKey),
		zap.Int("token_timeout", a.config.TokenTimeout),
		zap.Int("orders_queue_size", a.config.OrdersQueueSize),
		zap.Int("accural_workers", a.config.AccrualWorkers),
		zap.Int("accural_delayed_workers", a.config.AccrualDelayedWorkers),
		zap.Int("accural_delay", a.config.AccrualDelay),
		zap.Int("accural_delayed_batch", a.config.AccrualDelayedBatch),
		zap.Int("accural_req_repeats", a.config.AccrualRequestRepeats),
	)

	a.storage = postgresql.NewPsqlStorage(a.config.DatabaseUri)

	err := a.storage.InitStorage(ctx)
	if err != nil {
		return err
	}

	//make chan for transferring orders
	orderChan := make(chan *models.Order, a.config.OrdersQueueSize)

	//create accural client
	accrualClient := accrual.NewAccrualClient(
		a.config.AccrualSystemAddress,
		a.storage,
		a.config.AccrualWorkers,
		a.config.AccrualDelayedWorkers,
		a.config.AccrualDelay,
		a.config.AccrualDelayedBatch,
		orderChan,
		a.config.AccrualRequestRepeats)
	accrualClient.Start(ctx)

	// TODO переделать на получение из конфига
	authorizer := jwt.NewJwtTokenizer(a.config.TokenKey, time.Duration(a.config.TokenTimeout)*time.Hour)
	router := handlers.NewHttpRouter(a.storage, authorizer, orderChan)

	err = router.RouterInit(ctx)
	if err != nil {
		return err
	}
	err = router.StartRouter(a.config.RunAddress)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) Stop(cancel context.CancelFunc) {
	logger.Log.Debug("Syncing logger")
	logger.Log.Sync()
	a.storage.DbClose()
	cancel()
	//wait for logging from workers
	time.Sleep(time.Second * 1)
}
