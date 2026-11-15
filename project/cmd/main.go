package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TheTeaToper/OTUS/project/internal/app"
	"github.com/TheTeaToper/OTUS/project/internal/config"
	httpserver "github.com/TheTeaToper/OTUS/project/internal/server/http"
)

func main() {
	log.Println("Инициализация сервиса Анти-брутфорс...")

	// 1. Загружаем конфигурацию
	configPath := "configs/config.yaml"
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}
	log.Printf("Конфигурация успешно загружена из %s", configPath)

	// 2. Инициализируем бизнес-логику (сервис)
	appConfig := app.Config{
		LoginLimit:    cfg.Limits.LoginLimit,
		PasswordLimit: cfg.Limits.PasswordLimit,
		IpLimit:       cfg.Limits.IpLimit,
	}
	service := app.NewAntibruteforceService(appConfig, cfg.Limits.BucketTTL)
	defer service.Stop()

	// 3. Создаем HTTP сервер
	handler := httpserver.NewServer(service)
	server := &http.Server{
		Addr:         cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Канал для перехвата системных сигналов остановки (Ctrl+C, kill)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// 4. Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("HTTP сервер запущен на порту %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Ошибка работы HTTP сервера: %v", err)
		}
	}()

	// Ждем сигнал остановки
	<-stop
	log.Println("Получен сигнал остановки. Завершение работы...")

	// Плавное завершение работы (даем серверу 5 секунд на закрытие текущих соединений)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		cancel()
		log.Printf("Ошибка при плавном завершении работы сервера: %v", err)
	}
	cancel()

	log.Println("Сервис Анти-брутфорс успешно остановлен.")
}
