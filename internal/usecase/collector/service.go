package collector

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/Zmey56/poloniex-collector/internal/domain/models"
	"github.com/Zmey56/poloniex-collector/internal/domain/repository"
)

type Service struct {
	tradeRepo  repository.TradeRepository
	klineRepo  repository.KlineRepository
	exchange   repository.ExchangeClient
	workerPool *WorkerPool
}

type WorkerPool struct {
	numWorkers int
	taskQueue  chan *models.RecentTrade
	wg         sync.WaitGroup
}

func NewService(
	tradeRepo repository.TradeRepository,
	klineRepo repository.KlineRepository,
	exchange repository.ExchangeClient,
	numWorkers int,
) *Service {
	return &Service{
		tradeRepo:  tradeRepo,
		klineRepo:  klineRepo,
		exchange:   exchange,
		workerPool: NewWorkerPool(numWorkers),
	}
}

func NewWorkerPool(numWorkers int) *WorkerPool {
	return &WorkerPool{
		numWorkers: numWorkers,
		taskQueue:  make(chan *models.RecentTrade, 1000),
	}
}

func (s *Service) Run(ctx context.Context) error {
	// Запускаем воркеров
	s.workerPool.Start(ctx, s.processTradeWorker)

	// Список пар для мониторинга
	pairs := []string{"BTC_USDT", "ETH_USDT", "TRX_USDT", "DOGE_USDT", "BCH_USDT"}

	// Получаем исторические данные
	if err := s.loadHistoricalData(ctx, pairs); err != nil {
		return fmt.Errorf("load historical data error: %w", err)
	}

	// Подписываемся на текущие сделки
	tradeChan, err := s.exchange.SubscribeToTrades(ctx, pairs)
	if err != nil {
		return fmt.Errorf("subscribe to trades error: %w", err)
	}

	// Обрабатываем входящие сделки
	for trade := range tradeChan {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case s.workerPool.taskQueue <- &trade:
			// Сделка добавлена в очередь
		default:
			log.Printf("Warning: trade queue is full, skipping trade %s", trade.Tid)
		}
	}

	return nil
}

func (s *Service) loadHistoricalData(ctx context.Context, pairs []string) error {
	// Получаем данные за последние 24 часа
	endTime := time.Now().Unix()
	startTime := endTime - 86400 // 24 hours

	timeframes := []string{"1m", "15m", "1h", "1d"}

	for _, pair := range pairs {
		for _, timeframe := range timeframes {
			klines, err := s.exchange.GetHistoricalKlines(ctx, pair, timeframe, startTime, endTime)
			if err != nil {
				return fmt.Errorf("get historical klines error: %w", err)
			}

			for _, kline := range klines {
				if err := s.klineRepo.SaveKline(ctx, kline); err != nil {
					return fmt.Errorf("save kline error: %w", err)
				}
			}
		}
	}

	return nil
}

func (wp *WorkerPool) Start(ctx context.Context, processor func(context.Context, *models.RecentTrade) error) {
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go func() {
			defer wp.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case trade := <-wp.taskQueue:
					if err := processor(ctx, trade); err != nil {
						log.Printf("Error processing trade: %v", err)
					}
				}
			}
		}()
	}
}

func (s *Service) processTradeWorker(ctx context.Context, trade *models.RecentTrade) error {
	// Сохраняем сделку
	if err := s.tradeRepo.SaveTrade(ctx, *trade); err != nil {
		return fmt.Errorf("save trade error: %w", err)
	}

	// Обновляем клайны для всех таймфреймов
	timeframes := []string{"1m", "15m", "1h", "1d"}
	for _, tf := range timeframes {
		if err := s.updateKline(ctx, trade, tf); err != nil {
			return fmt.Errorf("update kline error: %w", err)
		}
	}

	return nil
}

func (s *Service) updateKline(ctx context.Context, trade *models.RecentTrade, timeframe string) error {
	interval := getTimeframeInterval(timeframe)
	beginTime := trade.Timestamp - (trade.Timestamp % interval)
	endTime := beginTime + interval

	// Получаем существующий клайн или создаем новый
	kline, err := s.klineRepo.GetLastKline(ctx, trade.Pair, timeframe)
	if err != nil {
		price, err := strconv.ParseFloat(trade.Price, 64)
		if err != nil {
			return fmt.Errorf("parse price error: %w", err)
		}

		kline = &models.Kline{
			Pair:      trade.Pair,
			TimeFrame: timeframe,
			O:         price,
			H:         price,
			L:         price,
			C:         price,
			UtcBegin:  beginTime,
			UtcEnd:    endTime,
			VolumeBS:  models.VBS{},
		}
	}

	// Обновляем данные клайна
	price, err := strconv.ParseFloat(trade.Price, 64)
	if err != nil {
		return fmt.Errorf("parse price error: %w", err)
	}

	amount, err := strconv.ParseFloat(trade.Amount, 64)
	if err != nil {
		return fmt.Errorf("parse amount error: %w", err)
	}

	kline.H = math.Max(kline.H, price)
	kline.L = math.Min(kline.L, price)
	kline.C = price

	if trade.Side == "buy" {
		kline.VolumeBS.BuyBase += amount
		kline.VolumeBS.BuyQuote += amount * price
	} else {
		kline.VolumeBS.SellBase += amount
		kline.VolumeBS.SellQuote += amount * price
	}

	return s.klineRepo.SaveKline(ctx, *kline)
}

func getTimeframeInterval(timeframe string) int64 {
	switch timeframe {
	case "1m":
		return 60
	case "15m":
		return 900
	case "1h":
		return 3600
	case "1d":
		return 86400
	default:
		return 60
	}
}
