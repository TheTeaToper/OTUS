package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/app"
	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/logger"
	internalhttp "github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/server/http"
	storage "github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/storage"
	inmemorystorage "github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/storage/inmemory"
	sqlstorage "github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config, err := LoadConfig(configFile)
	if err != nil {
		fmt.Printf("Failed to load cfg from %s with error: %v", configFile, err)
		config = NewConfig()
	}

	logger := logger.New(config.Logger.Level)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	var storage storage.Storage
	switch config.Storage.Type {
	case "in-memory":
		logger.Info("In-memory storage initialization...")
		storage = inmemorystorage.New()
		logger.Info("In-memory storage successfully initialized")
	case "sql":
		logger.Info("Sql storage initialization...")
		sqlStorage := sqlstorage.New(config.Storage.DSN, logger)
		connectCtx, connectCalcel := context.WithTimeout(ctx, 5*time.Second)
		if err := sqlStorage.Connect(connectCtx); err != nil {
			connectCalcel()
			logger.Error(fmt.Sprintf("Database connection error: %v", err))
			os.Exit(1) //nolint:gocritic
		}
		connectCalcel()
		logger.Info("Sql storage successfully initialized")
		defer func() {
			logger.Info("Closing database connection...")
			closeCtx, closeCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer closeCancel()
			if err := sqlStorage.Close(closeCtx); err != nil {
				logger.Error(fmt.Sprintf("Closing database connection error: %v", err))
			}
		}()
		storage = sqlStorage
	default:
		logger.Error(fmt.Sprintf("Incorrect storage type: %s", config.Storage.Type))
		os.Exit(1) //nolint:gocritic
	}

	calendar := app.New(logger, storage)

	server := internalhttp.NewServer(internalhttp.ServerConf(config.Server), calendar, storage, logger)

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logger.Error("failed to stop http server: " + err.Error())
		}
	}()

	logger.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		logger.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
