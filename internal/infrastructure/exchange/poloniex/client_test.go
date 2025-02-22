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
			Code int        `json:"code"`
			Data [][]string `json:"data"`
			Msg  string     `json:"msg"`
		}{
			Code: 200,
			Data: [][]string{
				{
					"58651",         // open
					"58651",         // high
					"58651",         // low
					"58651",         // close
					"1000",          // amount (quote currency)
					"500",           // quantity (base currency)
					"10",            // trade count
					"0",             // reserved
					"0",             // reserved
					"1719975420000", // start time
					"1719975479999", // end time
				},
			},
			Msg: "Success",
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
	assert.Equal(t, int64(1719975479), kline.UtcEnd)

	assert.Equal(t, 250.0, kline.VolumeBS.BuyBase)
	assert.Equal(t, 250.0, kline.VolumeBS.SellBase)
	assert.Equal(t, 500.0, kline.VolumeBS.BuyQuote)
	assert.Equal(t, 500.0, kline.VolumeBS.SellQuote)
}
