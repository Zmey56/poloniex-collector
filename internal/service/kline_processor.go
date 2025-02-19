package service

import (
	"context"

	"github.com/Zmey56/poloniex-collector/internal/domain/models"
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

func (p *KlineProcessor) ProcessTrade(ctx context.Context, trade *models.RecentTrade) error {
	// Логика обработки трейда и создания/обновления клайна
	// Эта функция должна использовать repository для сохранения данных
	return nil
}
