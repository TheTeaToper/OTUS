package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/queue"
	storage "github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/storage"
	sqlstorage "github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar_scheduler/config.yaml", "Path to configuration file")
}

//nolint:gocognit
func main() {
	flag.Parse()

	config, err := LoadConfig(configFile)
	if err != nil {
		fmt.Printf("Failed to load cfg from %s with error: %v", configFile, err)
		config = NewConfig()
	}
	if config.Storage.Type != "sql" {
		log.Fatal("Scheduler can`t work with inmemory storage of calendar service")
	}
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	logger := logger.New(config.Logger.Level)
	logger.Info("Sql storage initialization...")

	var storage storage.Storage
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

	amqpProducer, err := queue.NewProducer(config.Amqp.Url, config.Amqp.QueueName)
	if err != nil {
		log.Fatalf("failed to init amqp producer: %v", err)
	}
	defer amqpProducer.Close()

	logger.Info("Scheduler is running...")

	scanningTicker := time.NewTicker(time.Duration(config.ScanningIntervalInSeconds * int64(time.Second)))
	cleaningTicker := time.NewTicker(time.Duration(config.CleaningIntervalInHours * int64(time.Hour)))

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping scheduler...")
			return
		case <-scanningTicker.C:
			events, err := storage.GetEventsForNotification()
			if err != nil {
				logger.Error("failed to get events: %v", err)
				continue
			}

			for _, event := range events {
				if ctx.Err() != nil {
					break
				}
				body, err := json.Marshal(event)
				if err != nil {
					logger.Error("marshalling error: %v", err)
					continue
				}

				if err := amqpProducer.Publish(ctx, body); err != nil {
					logger.Error("publishing error: %v", err)
					continue
				}

				if err := storage.MarkEventAsNotified(event.ID); err != nil {
					logger.Error("failed to mark event as notified: %v", err)
				}
			}

		case <-cleaningTicker.C:
			rowsDeleted, err := storage.CleanOldEvents()
			if err != nil {
				logger.Error("failed to clean old events: %v", err)
			} else if rowsDeleted > 0 {
				logger.Info("Successfully cleaned events: %d", rowsDeleted)
			}
		}
	}
}
