package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/queue"
	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/jmoiron/sqlx"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar_sender/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	config, err := LoadConfig(configFile)
	if err != nil {
		fmt.Printf("Failed to load cfg from %s with error: %v", configFile, err)
		config = NewConfig()
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	logger := logger.New(config.Logger.Level)

	db, err := sqlx.Connect("pgx", config.DSN)
	if err != nil {
		logger.Error("failed to connect to db: %v", err)
		return
	}
	defer db.Close()

	amqpConsumer, err := queue.NewConsumer(config.Amqp.Url, config.Amqp.QueueName)
	if err != nil {
		logger.Error("failed to init amqp consumer: %v", err)
		return
	}
	defer amqpConsumer.Close()

	logger.Info("Sender is running...")

	handler := func(body []byte) error {
		var event storage.Event
		if err := json.Unmarshal(body, &event); err != nil {
			logger.Error("failed to parse event: %v", err)
			return err
		}

		logger.Info("[NOTIFICATION] User: %s! Event: '%s'. Starting time: %s",
			event.User, event.Title, event.StartTime.Format("2006-01-02 15:04:05"))
		_, err = db.ExecContext(ctx,
			"INSERT INTO notifications_log (user_id, event_title, status, sent_at) VALUES ($1, $2, $3, $4)",
			event.User, event.Title, "sent", time.Now())
		if err != nil {
			logger.Error("failed to log notification to db: %v", err)
		}
		return nil
	}

	if err := amqpConsumer.Consume(ctx, handler); err != nil {
		logger.Error("failed to start consumer: %v", err)
		return
	}

	<-ctx.Done()
	logger.Info("Stopping sender...")
}
