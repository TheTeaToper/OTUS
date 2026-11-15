.PHONY: build run stop test clean lint

# Переменные
BINARY_NAME=antibruteforce
CLI_NAME=ab-admin

# Сборка проекта локально
build:
	@echo "==> Сборка локальных бинарников..."
	go build -o $(BINARY_NAME) ./cmd/main.go
	go build -o $(CLI_NAME) ./cmd/cli/main.go

# Запуск проекта в Docker
run:
	@echo "==> Запуск сервиса в Docker Compose..."
	docker compose up --build -d

# Остановка контейнеров
stop:
	@echo "==> Остановка контейнеров Docker..."
	docker compose down

# Автоматический запуск тестов в правильном окружении
test:
	@echo "==> Поднимаем сервис в Docker для интеграционных тестов..."
	docker compose up --build -d
	@echo "==> Ожидаем инициализацию сервера (3 секунды)..."
	@sleep 3
	@echo "==> Запуск всех тестов (Unit + Integration)..."
	go test -v -race ./...
	@echo "==> Тесты завершены. Останавливаем контейнеры..."
	docker compose down

# Очистка собранных файлов
clean:
	@echo "==> Очистка временных файлов..."
	rm -f $(BINARY_NAME) $(CLI_NAME)
	docker compose down --v --rmi local 2>/dev/null || true

# Проверка кода линтером
lint:
	@echo "==> Запуск golangci-lint..."
	golangci-lint run ./...
