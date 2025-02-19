BINARY_NAME=poloniex-collector
MIGRATION_BINARY=migrator

.PHONY: all build test clean migrations generate mocks run docker-up docker-down

all: clean generate test build

build:
	go build -o bin/${BINARY_NAME} cmd/collector/main.go
	go build -o bin/${MIGRATION_BINARY} cmd/migrator/main.go

test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	go clean
	rm -f bin/${BINARY_NAME}
	rm -f bin/${MIGRATION_BINARY}
	rm -f coverage.out
	rm -f coverage.html

# Генерация моков
mocks:
	mockgen -destination=internal/domain/repository/mocks/mock_repository.go -package=mocks github.com/Zmey56/poloniex-collector/internal/domain/repository TradeRepository
	mockgen -destination=internal/domain/repository/mocks/mock_kline_repository.go -package=mocks github.com/Zmey56/poloniex-collector/internal/domain/repository KlineRepository
	mockgen -destination=internal/domain/repository/mocks/mock_exchange_client.go -package=mocks github.com/Zmey56/poloniex-collector/internal/domain/repository ExchangeClient

# Миграции
migrations-up:
	go run cmd/migrator/main.go "postgres://postgres:postgres@localhost:5432/poloniex?sslmode=disable" up

migrations-down:
	go run cmd/migrator/main.go "postgres://postgres:postgres@localhost:5432/poloniex?sslmode=disable" down

# Запуск приложения
run:
	go run cmd/collector/main.go

# Docker команды
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down