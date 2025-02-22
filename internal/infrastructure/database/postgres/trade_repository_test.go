package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Zmey56/poloniex-collector/internal/domain/models"
	"github.com/Zmey56/poloniex-collector/test/integration"
)

func TestTradeRepository_SaveTrade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	container, err := integration.NewPostgresContainer(t)
	require.NoError(t, err)
	defer container.Close()

	repo := NewTradeRepository(container.Pool)

	tests := []struct {
		name    string
		trade   models.RecentTrade
		wantErr bool
	}{
		{
			name: "valid trade",
			trade: models.RecentTrade{
				Tid:       "123",
				Pair:      "BTC_USDT",
				Price:     "50000.00",
				Amount:    "1.5",
				Side:      "buy",
				Timestamp: time.Now().Unix(),
			},
			wantErr: false,
		},
		{
			name: "duplicate trade",
			trade: models.RecentTrade{
				Tid:       "123",
				Pair:      "BTC_USDT",
				Price:     "50000.00",
				Amount:    "1.5",
				Side:      "buy",
				Timestamp: time.Now().Unix(),
			},
			wantErr: false, // должно игнорировать дубликаты
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.SaveTrade(context.Background(), tt.trade)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestTradeRepository_SaveTrades(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	container, err := integration.NewPostgresContainer(t)
	require.NoError(t, err)
	defer container.Close()

	repo := NewTradeRepository(container.Pool)

	trades := []models.RecentTrade{
		{
			Tid:       "123",
			Pair:      "BTC_USDT",
			Price:     "50000.00",
			Amount:    "1.5",
			Side:      "buy",
			Timestamp: time.Now().Unix(),
		},
		{
			Tid:       "124",
			Pair:      "BTC_USDT",
			Price:     "50100.00",
			Amount:    "2.0",
			Side:      "sell",
			Timestamp: time.Now().Unix(),
		},
	}

	err = repo.SaveTrades(context.Background(), trades)
	assert.NoError(t, err)
}
