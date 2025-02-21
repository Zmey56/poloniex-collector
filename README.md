# Poloniex Collector

## Описание
**Poloniex Collector** — это сервис для сбора и обработки рыночных данных с биржи Poloniex, включая сделки и свечные графики (klines). Проект использует PostgreSQL для хранения данных и предоставляет гибкую архитектуру для масштабирования.

## Структура проекта
```
poloniex-collector/
├── cmd/
│   ├── migrator/        # Сервис миграции базы данных
│   ├── collector/       # Основной сервис сбора данных
├── internal/
│   ├── config/          # Конфигурационные файлы
│   ├── service/         # Логика обработки данных
│   ├── infrastructure/  # Взаимодействие с базой данных, метриками и API Poloniex
│   │   ├── database/    # Работа с PostgreSQL
│   │   ├── metrics/     # Метрики и мониторинг
│   │   ├── exchange/    # Взаимодействие с биржей Poloniex
│   ├── usecase/         # Бизнес-логика
│   ├── repository/      # Репозитории для работы с данными
│   ├── domain/          # Определение бизнес-объектов (модели данных)
├── migrations/          # SQL-скрипты для инициализации базы
├── test/
│   ├── integration/     # Интеграционные тесты
│   ├── unit/            # Юнит-тесты
├── configs/             # Конфигурационные файлы (YAML, ENV)
├── Dockerfile           # Контейнеризация сервиса
├── docker-compose.yml   # Конфигурация окружения с базой данных
├── go.mod               # Управление зависимостями
├── go.sum               # Контроль версий зависимостей
├── Makefile             # Автоматизация сборки и запуска
├── README.md            # Документация проекта
├── .gitignore           # Исключение файлов из репозитория
├── UML_Diagram.png      # Диаграмма архитектуры
```

## Установка

### Требования
- **Go 1.22+**
- **PostgreSQL**
- **Docker & Docker Compose** (для контейнеризации)
- **Redis** (для кеширования)

### Запуск локально
1. Клонируйте репозиторий:
   ```sh
   git clone https://github.com/your-repo/poloniex-collector.git
   cd poloniex-collector
   ```
2. Установите зависимости:
   ```sh
   go mod tidy
   ```
3. Настройте переменные окружения (`.env` или `config.yaml`).
4. Запустите сервис миграции базы данных:
   ```sh
   go run cmd/migrator/main.go
   ```
5. Запустите основной сервис сбора данных:
   ```sh
   go run cmd/collector/main.go
   ```

## Конфигурация
Конфигурация проекта задается в файле `config.yaml`:
```yaml
database_url: "postgres://user:password@localhost:5432/poloniex"
exchange_api_key: "your_api_key"
exchange_api_secret: "your_api_secret"
log_level: "info"
cache_enabled: true
cache_ttl: 60 # Время жизни кеша в секундах
metrics_enabled: true
```

## Тестирование
Для запуска тестов используйте:
```sh
# Запуск всех тестов
go test ./...
```

Дополнительно можно проверить тесты с логами:
```sh
go test ./... -v
```

## Развертывание с Docker
Вы можете запустить сервис в контейнере:
```sh
docker-compose up --build
```

Если вам нужно запустить миграции перед стартом сервиса:
```sh
docker-compose run --rm migrator
```

