package poloniex

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetHistoricalKlines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/markets/BTC_USDT/candles", r.URL.Path)
		assert.Equal(t, "MINUTE_1", r.URL.Query().Get("interval"))

		response := struct {
			Candles []struct {
				Symbol     string `json:"symbol"`
				Interval   string `json:"interval"`
				Time       int64  `json:"time"`
				OpenPrice  string `json:"openPrice"`
				ClosePrice string `json:"closePrice"`
				HighPrice  string `json:"highPrice"`
				LowPrice   string `json:"lowPrice"`
				Volume     string `json:"volume"`
				Amount     string `json:"amount"`
			} `json:"candles"`
		}{
			Candles: []struct {
				Symbol     string `json:"symbol"`
				Interval   string `json:"interval"`
				Time       int64  `json:"time"`
				OpenPrice  string `json:"openPrice"`
				ClosePrice string `json:"closePrice"`
				HighPrice  string `json:"highPrice"`
				LowPrice   string `json:"lowPrice"`
				Volume     string `json:"volume"`
				Amount     string `json:"amount"`
			}{
				{
					Symbol:     "BTC_USDT",
					Interval:   "MINUTE_1",
					Time:       1719975420000,
					OpenPrice:  "58651",
					ClosePrice: "58651",
					HighPrice:  "58651",
					LowPrice:   "58651",
					Volume:     "500",
					Amount:     "1000",
				},
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(
		"ws://localhost",
		server.URL,
	)

	klines, err := client.GetHistoricalKlines(
		context.Background(),
		"BTC_USDT",
		"MINUTE_1",
		time.Now().Add(-time.Hour).Unix(),
		time.Now().Unix(),
	)

	require.NoError(t, err)
	require.Len(t, klines, 1)

	kline := klines[0]
	assert.Equal(t, "BTC_USDT", kline.Pair)
	assert.Equal(t, "MINUTE_1", kline.TimeFrame)
	assert.Equal(t, 58651.0, kline.O)
	assert.Equal(t, 58651.0, kline.H)
	assert.Equal(t, 58651.0, kline.L)
	assert.Equal(t, 58651.0, kline.C)
	assert.Equal(t, int64(1719975420), kline.UtcBegin)

	assert.Equal(t, 250.0, kline.VolumeBS.BuyBase)
	assert.Equal(t, 250.0, kline.VolumeBS.SellBase)
	assert.Equal(t, 500.0, kline.VolumeBS.BuyQuote)
	assert.Equal(t, 500.0, kline.VolumeBS.SellQuote)
}
