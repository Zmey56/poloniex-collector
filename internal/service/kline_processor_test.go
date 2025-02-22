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

func TestKlineProcessor_ProcessTrade(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockKlineRepository(ctrl)
	processor := NewKlineProcessor(mockRepo)

	tests := []struct {
		name      string
		trade     *models.RecentTrade
		setupMock func()
		wantErr   bool
	}{
		{
			name: "new kline",
			trade: &models.RecentTrade{
				Tid:       "123",
				Pair:      "BTC_USDT",
				Price:     "50000.00",
				Amount:    "1.5",
				Side:      "buy",
				Timestamp: time.Now().Unix(),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetLastKline(gomock.Any(), "BTC_USDT", "MINUTE_1").
					Return(nil, nil)

				mockRepo.EXPECT().
					SaveKline(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "update existing kline",
			trade: &models.RecentTrade{
				Tid:       "124",
				Pair:      "BTC_USDT",
				Price:     "51000.00",
				Amount:    "2.0",
				Side:      "sell",
				Timestamp: time.Now().Unix(),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetLastKline(gomock.Any(), "BTC_USDT", "MINUTE_1").
						Return(&models.Kline{
							Pair:      "BTC_USDT",
							TimeFrame: "MINUTE_1",
							O:         50000.0,
							H:         50000.0,
							L:         50000.0,
							C:         50000.0,
							UtcBegin:  time.Now().Unix(),
							UtcEnd:    time.Now().Add(time.Minute).Unix(),
							VolumeBS:  models.VBS{},
						}, nil)

				mockRepo.EXPECT().
					SaveKline(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := processor.ProcessTrade(context.Background(), tt.trade)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestKlineProcessor_MultipleTimeframes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockKlineRepository(ctrl)
	processor := NewKlineProcessor(mockRepo)

	trade := &models.RecentTrade{
		Tid:       "123",
		Pair:      "BTC_USDT",
		Price:     "50000.00",
		Amount:    "1.5",
		Side:      "buy",
		Timestamp: time.Now().Unix(),
	}

	timeframes := []string{"MINUTE_1", "MINUTE_15", "HOUR_1", "DAY_1"}

	for _, tf := range timeframes {
		mockRepo.EXPECT().
			GetLastKline(gomock.Any(), trade.Pair, tf).
			Return(nil, nil)

		mockRepo.EXPECT().
			SaveKline(gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, k models.Kline) {
					assert.Equal(t, tf, k.TimeFrame)
				}).
			Return(nil)
	}

	err := processor.ProcessTrade(context.Background(), trade)
	assert.NoError(t, err)
}
