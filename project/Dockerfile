# --- Этап 1: Сборка приложения ---
FROM golang:1.23-alpine AS builder

# Устанавливаем необходимые системные утилиты для сборки
RUN apk add --no-cache git make

WORKDIR /app

# Сначала копируем файлы зависимостей для кэширования слоев Docker
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код проекта
COPY . .

# Собираем сервер и CLI утилиту (статическая линковка, без CGO)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o antibruteforce ./cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o ab-admin ./cmd/cli/main.go

# --- Этап 2: Финальный легковесный образ ---
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /root/

# Копируем скомпилированные бинарники из этапа сборки
COPY --from=builder /app/antibruteforce .
COPY --from=builder /app/ab-admin .

# Копируем конфигурационные файлы
COPY --from=builder /app/configs ./configs

# Открываем порт, на котором работает сервер
EXPOSE 8080

# Команда по умолчанию для запуска контейнера
CMD ["./antibruteforce"]
