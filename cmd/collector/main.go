package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/Zmey56/poloniex-collector/internal/config"
	"github.com/Zmey56/poloniex-collector/internal/infrastructure/database/postgres"
	"github.com/Zmey56/poloniex-collector/internal/infrastructure/exchange/poloniex"
	"github.com/Zmey56/poloniex-collector/internal/usecase/collector"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключаемся к базе данных
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Failed to parse database config: %v", err)
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Создаем репозитории
	tradeRepo := postgres.NewTradeRepository(pool)
	klineRepo := postgres.NewKlineRepository(pool)

	// Создаем клиент Poloniex
	exchange := poloniex.NewClient(cfg.Poloniex.WSURL, cfg.Poloniex.RestURL)

	// Создаем сервис
	service := collector.NewService(
		tradeRepo,
		klineRepo,
		exchange,
		cfg.Worker.PoolSize,
	)

	// Создаем контекст с отменой
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обрабатываем сигналы завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	// Запускаем сервис
	log.Println("Starting service...")
	if err := service.Run(ctx); err != nil {
		log.Fatalf("Service error: %v", err)
	}
}
