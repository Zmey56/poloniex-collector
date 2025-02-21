package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/Zmey56/poloniex-collector/internal/domain/models"
	"github.com/Zmey56/poloniex-collector/test/mocks"
)

func TestWorkerPool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockKlineRepo := mocks.NewMockKlineRepository(ctrl)
	processor := NewKlineProcessor(mockKlineRepo)
	pool := NewWorkerPool(2, processor)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Подготавливаем тестовые данные
	trade := &models.RecentTrade{
		Tid:       123,
		Pair:      "BTC_USDT",
		Price:     50000.00,
		Amount:    1.5,
		Side:      "buy",
		Timestamp: time.Now().Unix(),
	}

	// Настраиваем ожидания для мока
	mockKlineRepo.EXPECT().
		GetLastKline(gomock.Any(), trade.Pair, "1m").
		Return(nil, nil).
		AnyTimes()

	mockKlineRepo.EXPECT().
		SaveKline(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	// Запускаем пул
	pool.Start(ctx)

	// Тестируем отправку задачи
	success := pool.Submit(trade)
	assert.True(t, success, "Should successfully submit trade")

	// Даем время на обработку
	time.Sleep(100 * time.Millisecond)

	// Тестируем переполнение очереди
	for i := 0; i < 1100; i++ { // Больше чем размер буфера
		pool.Submit(trade)
	}

	// Проверяем graceful shutdown
	cancel()
	pool.Stop()
}
