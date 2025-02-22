package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Zmey56/poloniex-collector/internal/domain/models"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetKlineByInterval(ctx context.Context, pair string, timeframe string, beginTime int64) (*models.Kline, error) {
	args := m.Called(ctx, pair, timeframe, beginTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Kline), args.Error(1)
}

func (m *MockRepository) GetLastKline(ctx context.Context, pair, timeframe string) (*models.Kline, error) {
	args := m.Called(ctx, pair, timeframe)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Kline), args.Error(1)
}

func (m *MockRepository) SaveKline(ctx context.Context, kline models.Kline) error {
	args := m.Called(ctx, kline)
	return args.Error(0)
}

func TestProcessTrade(t *testing.T) {
	t.Run("ProcessTrade_ValidData_CreatesNewKlines", func(t *testing.T) {
		mockRepo := new(MockRepository)
		processor := NewKlineProcessor(mockRepo)
		ctx := context.Background()

		trade := &models.RecentTrade{
			Pair:      "BTC_USDT",
			Price:     "50000.0",
			Amount:    "1.5",
			Side:      "buy",
			Timestamp: time.Now().Unix() * 1000,
		}

		for _, tf := range []string{
			PoloniexTimeFrame1m,
			PoloniexTimeFrame15m,
			PoloniexTimeFrame1h,
			PoloniexTimeFrame1d,
		} {
			mockRepo.On("GetLastKline", ctx, trade.Pair, tf).Return(nil, sql.ErrNoRows)
			mockRepo.On("SaveKline", ctx, mock.AnythingOfType("models.Kline")).Return(nil)
		}

		err := processor.ProcessTrade(ctx, trade)

		assert.NoError(t, err)
		mockRepo.AssertNumberOfCalls(t, "GetLastKline", 4) // По одному на каждый таймфрейм
		mockRepo.AssertNumberOfCalls(t, "SaveKline", 4)    // По одному на каждый таймфрейм
	})

	t.Run("ProcessTrade_ValidData_UpdatesExistingKlines", func(t *testing.T) {
		mockRepo := new(MockRepository)
		processor := NewKlineProcessor(mockRepo)
		ctx := context.Background()

		now := time.Now().UTC()
		beginTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.UTC)
		endTime := beginTime.Add(1 * time.Minute)

		trade := &models.RecentTrade{
			Pair:      "BTC_USDT",
			Price:     "50000.0",
			Amount:    "1.5",
			Side:      "sell",
			Timestamp: now.Unix() * 1000, // Текущее время в мс
		}

		existingKline := &models.Kline{
			Pair:      trade.Pair,
			TimeFrame: PoloniexTimeFrame1m,
			O:         49000.0,
			H:         49500.0,
			L:         48900.0,
			C:         49200.0,
			UtcBegin:  beginTime.Unix() * 1000,
			UtcEnd:    endTime.Unix() * 1000,
			BeginDt:   beginTime,
			EndDt:     endTime,
			VolumeBS: models.VBS{
				BuyBase:   2.0,
				SellBase:  1.0,
				BuyQuote:  98000.0,
				SellQuote: 49200.0,
			},
		}

		mockRepo.On("GetLastKline", ctx, trade.Pair, PoloniexTimeFrame1m).Return(existingKline, nil)
		mockRepo.On("SaveKline", ctx, mock.AnythingOfType("models.Kline")).Return(nil)

		for _, tf := range []string{
			PoloniexTimeFrame15m,
			PoloniexTimeFrame1h,
			PoloniexTimeFrame1d,
		} {
			mockRepo.On("GetLastKline", ctx, trade.Pair, tf).Return(nil, sql.ErrNoRows)
			mockRepo.On("SaveKline", ctx, mock.AnythingOfType("models.Kline")).Return(nil)
		}

		err := processor.ProcessTrade(ctx, trade)

		assert.NoError(t, err)
		mockRepo.AssertCalled(t, "GetLastKline", ctx, trade.Pair, PoloniexTimeFrame1m)
		mockRepo.AssertCalled(t, "SaveKline", ctx, mock.MatchedBy(func(k models.Kline) bool {
			expectedPrice := 50000.0
			expectedH := 50000.0
			expectedL := 48900.0

			assert.Equal(t, trade.Pair, k.Pair)
			assert.Equal(t, PoloniexTimeFrame1m, k.TimeFrame)
			assert.Equal(t, expectedPrice, k.C)
			assert.Equal(t, expectedH, k.H)
			assert.Equal(t, expectedL, k.L)

			// Проверяем объемы
			expectedSellBase := 2.5
			expectedSellQuote := 124200.0
			assert.Equal(t, expectedSellBase, k.VolumeBS.SellBase)
			assert.Equal(t, expectedSellQuote, k.VolumeBS.SellQuote)

			return true
		}))
	})

	t.Run("ProcessTrade_InvalidPrice_ReturnsError", func(t *testing.T) {
		mockRepo := new(MockRepository)
		processor := NewKlineProcessor(mockRepo)
		ctx := context.Background()

		trade := &models.RecentTrade{
			Pair:      "BTC_USDT",
			Price:     "invalid",
			Amount:    "1.5",
			Side:      "buy",
			Timestamp: time.Now().Unix() * 1000,
		}

		err := processor.ProcessTrade(ctx, trade)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid price format")
		mockRepo.AssertNotCalled(t, "SaveKline")
	})

	t.Run("ProcessTrade_InvalidAmount_ReturnsError", func(t *testing.T) {
		mockRepo := new(MockRepository)
		processor := NewKlineProcessor(mockRepo)
		ctx := context.Background()

		trade := &models.RecentTrade{
			Pair:      "BTC_USDT",
			Price:     "50000.0",
			Amount:    "invalid",
			Side:      "buy",
			Timestamp: time.Now().Unix() * 1000,
		}

		err := processor.ProcessTrade(ctx, trade)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid amount format")
		mockRepo.AssertNotCalled(t, "SaveKline")
	})

	t.Run("ProcessTrade_DatabaseError_ReturnsError", func(t *testing.T) {
		mockRepo := new(MockRepository)
		processor := NewKlineProcessor(mockRepo)
		ctx := context.Background()

		trade := &models.RecentTrade{
			Pair:      "BTC_USDT",
			Price:     "50000.0",
			Amount:    "1.5",
			Side:      "buy",
			Timestamp: time.Now().Unix() * 1000,
		}

		dbError := errors.New("database error")
		mockRepo.On("GetLastKline", ctx, trade.Pair, PoloniexTimeFrame1m).Return(nil, dbError)

		err := processor.ProcessTrade(ctx, trade)

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		mockRepo.AssertNotCalled(t, "SaveKline")
	})

	t.Run("ProcessTrade_SaveError_ReturnsError", func(t *testing.T) {

		mockRepo := new(MockRepository)
		processor := NewKlineProcessor(mockRepo)
		ctx := context.Background()

		trade := &models.RecentTrade{
			Pair:      "BTC_USDT",
			Price:     "50000.0",
			Amount:    "1.5",
			Side:      "buy",
			Timestamp: time.Now().Unix() * 1000,
		}

		mockRepo.On("GetLastKline", ctx, trade.Pair, PoloniexTimeFrame1m).Return(nil, sql.ErrNoRows)
		saveError := errors.New("save error")
		mockRepo.On("SaveKline", ctx, mock.AnythingOfType("models.Kline")).Return(saveError)

		err := processor.ProcessTrade(ctx, trade)

		assert.Error(t, err)
		assert.Equal(t, saveError, err)
	})
}

func TestGetKlineTimestamps(t *testing.T) {
	testCases := []struct {
		name          string
		timestamp     int64
		timeFrame     string
		expectedBegin int64
		expectedEnd   int64
	}{
		{
			name:          "1 minute timeframe with millisecond timestamp",
			timestamp:     1676548234000,
			timeFrame:     TimeFrame1m,
			expectedBegin: 1676548200000,
			expectedEnd:   1676548260000,
		},
		{
			name:          "15 minute timeframe with second timestamp",
			timestamp:     1676548234,
			timeFrame:     TimeFrame15m,
			expectedBegin: 1676547900000,
			expectedEnd:   1676548800000,
		},
		{
			name:          "1 hour timeframe with millisecond timestamp",
			timestamp:     1676548234000,
			timeFrame:     TimeFrame1h,
			expectedBegin: 1676545200000,
			expectedEnd:   1676548800000,
		},
		{
			name:          "1 day timeframe with second timestamp",
			timestamp:     1676548234,
			timeFrame:     TimeFrame1d,
			expectedBegin: 1676505600000,
			expectedEnd:   1676592000000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			beginTime, endTime := getKlineTimestamps(tc.timestamp, tc.timeFrame)

			assert.Equal(t, tc.expectedBegin, beginTime)
			assert.Equal(t, tc.expectedEnd, endTime)
		})
	}
}
