package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/madcarpet/gophermart/internal/app"
	"github.com/madcarpet/gophermart/internal/config"
	"github.com/madcarpet/gophermart/internal/logger"
	"go.uber.org/zap"
)

func main() {
	// Channels for signals
	osSigCh := make(chan os.Signal, 1)
	signal.Notify(osSigCh, syscall.SIGINT, syscall.SIGTERM)
	errCh := make(chan error)

	//Creating main ctx
	ctx, cancel := context.WithCancel(context.Background())

	// Config initialization
	appCfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("application config initialisation failed err: %v", err)
	}

	// App initialization
	app := app.NewApp(*appCfg)
	// App starting with configuration
	go func() {
		err = app.Start(ctx)
		if err != nil {
			log.Fatalf("application start failed err: %v", err)
		}
	}()

	select {
	case sig := <-osSigCh:
		logger.Log.Info("Stopping application, os sig received", zap.String("signal", sig.String()))
		app.Stop(cancel)
	case err := <-errCh:
		logger.Log.Error("Application error", zap.Error(err))
		app.Stop(cancel)
	}
}
