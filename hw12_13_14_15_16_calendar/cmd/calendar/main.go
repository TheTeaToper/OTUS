package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/app"
	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/logger"
	internalgrpc "github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/server/grpc"
	internalhttp "github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/server/http"
	storage "github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/storage"
	inmemorystorage "github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/storage/inmemory"
	sqlstorage "github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/storage/sql"
	"golang.org/x/sync/errgroup"
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
			logger.Error("Database connection error: %v", err)
			os.Exit(1) //nolint:gocritic
		}
		connectCalcel()
		logger.Info("Sql storage successfully initialized")
		defer func() {
			logger.Info("Closing database connection...")
			closeCtx, closeCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer closeCancel()
			if err := sqlStorage.Close(closeCtx); err != nil {
				logger.Error("Closing database connection error: %v", err)
			}
		}()
		storage = sqlStorage
	default:
		logger.Error("Incorrect storage type: %s", config.Storage.Type)
		os.Exit(1)
	}

	calendar := app.New(logger, storage)

	httpServer := internalhttp.NewServer(internalhttp.ServerConf(config.Server), calendar, logger)

	// адрес для gRPC-сервера (Host:Port+1)
	basePort, err := strconv.Atoi(config.Server.Port)
	if err != nil {
		logger.Error("Invalid port in config: %v", err)
		os.Exit(1)
	}
	grpcAddr := net.JoinHostPort(config.Server.Host, strconv.Itoa(basePort+1))
	grpcServer := internalgrpc.NewServer(calendar, logger)

	g, groupCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		go func() {
			<-ctx.Done()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			if err := httpServer.Stop(ctx); err != nil {
				logger.Error("failed to stop http server: %v", err.Error())
			}
		}()
		return httpServer.Start(groupCtx)
	})

	g.Go(func() error {
		return grpcServer.Start(groupCtx, grpcAddr)
	})

	logger.Info("calendar is running...")

	if err := g.Wait(); err != nil && groupCtx.Err() == nil {
		logger.Error("server error: %v", err)
		cancel()
		os.Exit(1)
	}

	logger.Info("calendar stopped successfully")
}
