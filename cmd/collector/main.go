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
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("Starting Poloniex collector...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	pool, err := pgxpool.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	tradeRepo := postgres.NewTradeRepository(pool)
	klineRepo := postgres.NewKlineRepository(pool)

	exchange := poloniex.NewClient(cfg.Poloniex.WSURL, cfg.Poloniex.RestURL)

	service := collector.NewService(
		tradeRepo,
		klineRepo,
		exchange,
		cfg.Worker.PoolSize,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		if err := service.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-sigChan:
		log.Println("Received shutdown signal")
		cancel()
	case err := <-errChan:
		log.Printf("Service error: %v", err)
		cancel()
	}

	log.Println("Shutdown complete")
}
