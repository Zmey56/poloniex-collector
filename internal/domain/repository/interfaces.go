package repository

import (
	"context"

	"github.com/Zmey56/poloniex-collector/internal/domain/models"
)

type TradeRepository interface {
	SaveTrade(ctx context.Context, trade models.RecentTrade) error
	SaveTrades(ctx context.Context, trades []models.RecentTrade) error
}

type KlineRepository interface {
	SaveKline(ctx context.Context, kline models.Kline) error
	GetLastKline(ctx context.Context, pair, timeframe string) (*models.Kline, error)
	GetKlinesByTimeRange(ctx context.Context, pair, timeframe string, startTime, endTime int64) ([]models.Kline, error)
	GetKlineByInterval(ctx context.Context, pair, timeframe string, beginTime int64) (*models.Kline, error)
}

type ExchangeClient interface {
	GetHistoricalKlines(ctx context.Context, pair string, timeframe string, startTime, endTime int64) ([]models.Kline, error)
	SubscribeToTrades(ctx context.Context, pairs []string) (<-chan models.RecentTrade, error)
}
