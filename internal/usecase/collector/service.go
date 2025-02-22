package collector

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Zmey56/poloniex-collector/internal/domain/repository"
	"github.com/Zmey56/poloniex-collector/internal/service"
)

type Service struct {
	tradeRepo  repository.TradeRepository
	klineRepo  repository.KlineRepository
	exchange   repository.ExchangeClient
	workerPool *service.WorkerPool
}

func NewService(
	tradeRepo repository.TradeRepository,
	klineRepo repository.KlineRepository,
	exchange repository.ExchangeClient,
	numWorkers int,
) *Service {
	// Создаем процессор для обработки клайнов
	klineProcessor := service.NewKlineProcessor(klineRepo)

	// Создаем пул воркеров с процессором
	workerPool := service.NewWorkerPool(numWorkers, klineProcessor)

	return &Service{
		tradeRepo:  tradeRepo,
		klineRepo:  klineRepo,
		exchange:   exchange,
		workerPool: workerPool,
	}
}

func (s *Service) Run(ctx context.Context) error {
	// Запускаем пул воркеров
	s.workerPool.Start(ctx)
	log.Println("Worker pool started")

	// Список торговых пар
	pairs := []string{"BTC_USDT", "ETH_USDT", "TRX_USDT", "DOGE_USDT", "BCH_USDT"}

	// Загружаем исторические данные
	if err := s.loadHistoricalData(ctx, pairs); err != nil {
		return fmt.Errorf("load historical data error: %w", err)
	}

	// Подписываемся на трейды
	trades, err := s.exchange.SubscribeToTrades(ctx, pairs)
	if err != nil {
		return fmt.Errorf("subscribe to trades error: %w", err)
	}

	// Обрабатываем входящие трейды
	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping service")
			s.workerPool.Stop()
			return ctx.Err()
		case trade, ok := <-trades:
			if !ok {
				return fmt.Errorf("trade channel closed")
			}

			// Сохраняем трейд
			if err := s.tradeRepo.SaveTrade(ctx, trade); err != nil {
				log.Printf("Error saving trade: %v", err)
				continue
			}

			// Отправляем трейд в пул воркеров для создания клайнов
			if ok := s.workerPool.Submit(&trade); !ok {
				log.Printf("Failed to submit trade to worker pool: queue is full")
			}
		}
	}
}

func (s *Service) loadHistoricalData(ctx context.Context, pairs []string) error {
	log.Println("Loading historical data...")
	timeframes := []string{"MINUTE_1", "MINUTE_15", "HOUR_1", "DAY_1"}
	endTime := time.Now().Unix()

	for _, pair := range pairs {
		for _, timeframe := range timeframes {
			lastKline, err := s.klineRepo.GetLastKline(ctx, pair, timeframe)

			var startTime int64
			if err == nil && lastKline != nil {
				startTime = lastKline.UtcEnd
				log.Printf("Found last kline for %s %s at %v, continuing from there",
					pair, timeframe, startTime)
			} else {
				// Если нет данных, начинаем с 1 декабря 2024
				startTime = 1701388800 // 2024-12-01 00:00:00 UTC
				log.Printf("No previous klines found for %s %s, starting from 2024-12-01",
					pair, timeframe)
			}

			if startTime >= endTime {
				log.Printf("No new data for %s %s", pair, timeframe)
				continue
			}

			klines, err := s.exchange.GetHistoricalKlines(ctx, pair, timeframe, startTime, endTime)
			if err != nil {
				return fmt.Errorf("get historical klines error: %w", err)
			}

			log.Printf("Received %d klines for %s %s", len(klines), pair, timeframe)

			for _, kline := range klines {
				log.Printf("Saving get historical klines: %+v", kline)
				if err := s.klineRepo.SaveKline(ctx, kline); err != nil {
					return fmt.Errorf("save kline error: %w", err)
				}
			}
		}
	}

	log.Println("Historical data loaded successfully")
	return nil
}
