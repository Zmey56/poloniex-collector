package service

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/Zmey56/poloniex-collector/internal/domain/models"
)

const (
	timeFrame1m  = "1m"
	timeFrame15m = "15m"
	timeFrame1h  = "1h"
	timeFrame1d  = "1d"
)

type KlineRepository interface {
	SaveKline(ctx context.Context, kline models.Kline) error
	GetLastKline(ctx context.Context, pair, timeframe string) (*models.Kline, error)
}

type KlineProcessor struct {
	repository KlineRepository
}

func NewKlineProcessor(repository KlineRepository) *KlineProcessor {
	return &KlineProcessor{
		repository: repository,
	}
}

type KlineAggregator struct {
	pair       string            // Валютная пара (BTC_USDT и т.д.)
	timeFrames []string          // Список поддерживаемых таймфреймов (1m, 15m, 1h, 1d)
	klines     map[string]*Kline // Активные свечи с ключом "timeFrame-timestamp"
	db         *sql.DB           // Соединение с базой данных
	mu         sync.Mutex        // Мьютекс для потокобезопасного доступа
}

type Kline struct {
	Pair      string  // название валютной пары (BTC_USDT и т.д.)
	TimeFrame string  // период формирования свечи (1m, 15m, 1h, 1d)
	O         float64 // open - цена открытия
	H         float64 // high - максимальная цена
	L         float64 // low - минимальная цена
	C         float64 // close - цена закрытия
	UtcBegin  int64   // время unix начала формирования свечки (в наносекундах)
	UtcEnd    int64   // время unix окончания формирования свечки (в наносекундах)
	VolumeBS  VBS     // объемы торгов с разделением на buy/sell
}

type VBS struct {
	BuyBase   float64 // объём покупок в базовой валюте
	SellBase  float64 // объём продаж в базовой валюте
	BuyQuote  float64 // объём покупок в котируемой валюте
	SellQuote float64 // объём продаж в котируемой валюте
}

func (p *KlineProcessor) ProcessTrade(ctx context.Context, trade *models.RecentTrade) error {
	// Use consistent timeframe format
	timeframes := []string{"MINUTE_1", "MINUTE_15", "HOUR_1", "DAY_1"}

	for _, timeframe := range timeframes {
		lastKline, err := p.repository.GetLastKline(ctx, trade.Pair, timeframe)
		if err != nil {
			return err
		}

		if lastKline == nil {
			newKline := models.Kline{
				Pair:      trade.Pair,
				TimeFrame: timeframe,
				O:         trade.Price,
				H:         trade.Price,
				L:         trade.Price,
				C:         trade.Price,
				UtcBegin:  trade.Timestamp,
				UtcEnd:    trade.Timestamp + getTimeframeSeconds(timeframe),
				VolumeBS: models.VBS{
					BuyBase: trade.Amount,
				},
			}
			if err := p.repository.SaveKline(ctx, newKline); err != nil {
				return err
			}
		} else {
			lastKline.H = max(lastKline.H, trade.Price)
			lastKline.L = min(lastKline.L, trade.Price)
			lastKline.C = trade.Price
			lastKline.VolumeBS.BuyBase += trade.Amount
			lastKline.VolumeBS.BuyQuote += trade.Amount * trade.Price
			lastKline.VolumeBS.SellBase += trade.Amount
			lastKline.VolumeBS.SellQuote += trade.Amount * trade.Price

			if err := p.repository.SaveKline(ctx, *lastKline); err != nil {
				return err
			}
		}
	}
	return nil
}

func getTimeframeSeconds(timeframe string) int64 {
	switch timeframe {
	case "MINUTE_1":
		return 60
	case "MINUTE_15":
		return 900
	case "HOUR_1":
		return 3600
	case "DAY_1":
		return 86400
	default:
		return 60
	}
}

func getKlineTimestamps(timestamp int64, timeFrame string) (int64, int64) {
	t := time.Unix(0, timestamp)

	var beginTime time.Time
	var endTime time.Time

	switch timeFrame {
	case timeFrame1m:
		beginTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)
		endTime = beginTime.Add(1 * time.Minute)
	case timeFrame15m:
		minute := t.Minute() / 15 * 15 // Округляем до ближайших 15 минут (0, 15, 30, 45)
		beginTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), minute, 0, 0, time.UTC)
		endTime = beginTime.Add(15 * time.Minute)
	case timeFrame1h:
		beginTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.UTC)
		endTime = beginTime.Add(1 * time.Hour)
	case timeFrame1d:
		beginTime = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		endTime = beginTime.Add(24 * time.Hour)
	}

	return beginTime.UnixNano(), endTime.UnixNano()
}
