package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Zmey56/poloniex-collector/internal/domain/models"
	"github.com/Zmey56/poloniex-collector/test/integration"
)

func TestKlineRepository_SaveAndGetKline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	container, err := integration.NewPostgresContainer(t)
	require.NoError(t, err)
	defer container.Close()

	repo := NewKlineRepository(container.Pool)

	kline := models.Kline{
		Pair:      "BTC_USDT",
		TimeFrame: "1m",
		O:         50000.0,
		H:         51000.0,
		L:         49000.0,
		C:         50500.0,
		UtcBegin:  time.Now().Truncate(time.Second).Unix(),
		UtcEnd:    time.Now().Add(time.Minute).Truncate(time.Second).Unix(),
		VolumeBS: models.VBS{
			BuyBase:   1.5,
			SellBase:  2.0,
			BuyQuote:  75000.0,
			SellQuote: 100000.0,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Сохраняем клайн
	err = repo.SaveKline(ctx, kline)
	require.NoError(t, err)

	// Ждем завершения транзакции перед чтением
	time.Sleep(1000 * time.Millisecond)

	// Получаем сохраненный клайн
	savedKline, err := repo.GetLastKline(ctx, kline.Pair, kline.TimeFrame)
	require.NoError(t, err)
	require.NotNil(t, savedKline)

	// Проверяем соответствие данных
	assert.Equal(t, kline.Pair, savedKline.Pair)
	assert.Equal(t, kline.TimeFrame, savedKline.TimeFrame)
	assert.Equal(t, kline.O, savedKline.O)
	assert.Equal(t, kline.H, savedKline.H)
	assert.Equal(t, kline.L, savedKline.L)
	assert.Equal(t, kline.C, savedKline.C)
	assert.Equal(t, kline.VolumeBS, savedKline.VolumeBS)
}

func TestKlineRepository_GetKlinesByTimeRange(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	container, err := integration.NewPostgresContainer(t)
	require.NoError(t, err)
	defer container.Close()

	repo := NewKlineRepository(container.Pool)

	// Создаем несколько клайнов с разным временем
	// firstTime := time.Now().Truncate(time.Second)

	now := time.Now().Truncate(time.Second)
	nowDiffMinute := now.Add(-time.Minute)
	nowDiffTwoMinutes := now.Add(-2 * time.Minute)
	klines := []models.Kline{
		createTestKline("BTC_USDT", "1m", nowDiffTwoMinutes),
		createTestKline("BTC_USDT", "1m", nowDiffMinute),
		createTestKline("BTC_USDT", "1m", now),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Сохраняем клайны
	for _, k := range klines {
		fmt.Printf("Saving kline: %+v\n", k)
		err := repo.SaveKline(ctx, k)
		require.NoError(t, err)
	}

	// Подсчитываем количество записей в базе данных
	var count int
	count, err = repo.CountKlines(ctx)
	require.NoError(t, err)
	fmt.Println("Count of klines: ", count)

	// Ждем завершения всех транзакций перед чтением
	time.Sleep(1000 * time.Millisecond)

	// Получаем клайны за определенный период
	startTime := nowDiffTwoMinutes.Unix() - 500 // 500 - чтобы попасть в диапазон
	endTime := time.Now().Add(time.Minute).Truncate(time.Second).Unix()
	fmt.Printf("Start time: %v, end time: %v\n", startTime, endTime)
	result, err := repo.GetKlinesByTimeRange(ctx, "BTC_USDT", "1m", startTime, endTime)
	require.NoError(t, err)
	for _, k := range result {
		fmt.Printf("Result kline: %+v\n", k)
	}

	// Проверяем, что получены 3 клайна
	require.Len(t, result, 3)
	assert.Len(t, result, 3)
}

func createTestKline(pair, timeframe string, timestamp time.Time) models.Kline {
	return models.Kline{
		Pair:      pair,
		TimeFrame: timeframe,
		O:         50000.0,
		H:         51000.0,
		L:         49000.0,
		C:         50500.0,
		UtcBegin:  timestamp.Unix(),
		UtcEnd:    timestamp.Add(time.Minute).Unix(),
		VolumeBS: models.VBS{
			BuyBase:   1.5,
			SellBase:  2.0,
			BuyQuote:  75000.0,
			SellQuote: 100000.0,
		},
	}
}
