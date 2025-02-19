# Dockerfile
FROM --platform=linux/amd64 golang:1.23.4-alpine AS builder

WORKDIR /app

# Обновляем индекс пакетов и устанавливаем необходимые зависимости
RUN apk update && \
    apk add --no-cache \
    gcc \
    musl-dev \
    make

# Копируем файлы с зависимостями
COPY go.mod .
COPY go.sum .

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение с указанием платформы
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o /app/bin/collector cmd/collector/main.go

# Финальный этап
FROM --platform=linux/amd64 alpine:latest

WORKDIR /app

# Копируем бинарный файл из builder
COPY --from=builder /app/bin/collector .
COPY internal/config/config.go ./config/

# Запускаем приложение
CMD ["./collector"]