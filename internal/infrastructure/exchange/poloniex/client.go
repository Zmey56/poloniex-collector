package poloniex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"

	"github.com/Zmey56/poloniex-collector/internal/domain/models"
)

type Client struct {
	wsURL   string
	restURL string
	client  *http.Client
}

func NewClient(wsURL, restURL string) *Client {
	return &Client{
		wsURL:   wsURL,
		restURL: restURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) GetHistoricalKlines(ctx context.Context, pair string, timeframe string, startTime, endTime int64) ([]models.Kline, error) {
	u := fmt.Sprintf("%s/markets/%s/candles?interval=%s&startTime=%d&endTime=%d",
		c.restURL,
		pair,
		timeframe,
		startTime*1000,
		endTime*1000)

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	var response struct {
		Data []struct {
			Open      string `json:"open"`
			High      string `json:"high"`
			Low       string `json:"low"`
			Close     string `json:"close"`
			Volume    string `json:"volume"`
			Timestamp int64  `json:"timestamp"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	klines := make([]models.Kline, len(response.Data))
	for i, candle := range response.Data {
		open, _ := strconv.ParseFloat(candle.Open, 64)
		high, _ := strconv.ParseFloat(candle.High, 64)
		low, _ := strconv.ParseFloat(candle.Low, 64)
		close, _ := strconv.ParseFloat(candle.Close, 64)
		volume, _ := strconv.ParseFloat(candle.Volume, 64)

		klines[i] = models.Kline{
			Pair:      pair,
			TimeFrame: timeframe,
			O:         open,
			H:         high,
			L:         low,
			C:         close,
			UtcBegin:  candle.Timestamp / 1000,
			UtcEnd:    (candle.Timestamp / 1000) + getTimeframeInterval(timeframe),
			VolumeBS: models.VBS{
				BuyBase:   volume / 2,
				SellBase:  volume / 2,
				BuyQuote:  (volume / 2) * ((open + close) / 2),
				SellQuote: (volume / 2) * ((open + close) / 2),
			},
		}
	}

	return klines, nil
}

func (c *Client) SubscribeToTrades(ctx context.Context, pairs []string) (<-chan models.RecentTrade, error) {
	trades := make(chan models.RecentTrade, 1000)

	u, err := url.Parse(c.wsURL)
	if err != nil {
		return nil, fmt.Errorf("parse ws url error: %w", err)
	}

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("dial ws error: %w", err)
	}

	// Subscribe to channels
	for _, pair := range pairs {
		subscribeMsg := map[string]interface{}{
			"event":   "subscribe",
			"channel": "trades",
			"symbols": []string{pair},
		}
		if err := conn.WriteJSON(subscribeMsg); err != nil {
			conn.Close()
			return nil, fmt.Errorf("subscribe error: %w", err)
		}
	}

	go func() {
		defer conn.Close()
		defer close(trades)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				var msg struct {
					Event   string `json:"event"`
					Channel string `json:"channel"`
					Data    []struct {
						Symbol   string `json:"symbol"`
						TradeID  string `json:"tradeId"`
						Price    string `json:"price"`
						Quantity string `json:"quantity"`
						Side     string `json:"side"`
						Ts       int64  `json:"ts"`
					} `json:"data"`
				}

				if err := conn.ReadJSON(&msg); err != nil {
					if websocket.IsUnexpectedCloseError(err) {
						// Try to reconnect
						time.Sleep(time.Second)
						if newConn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil); err == nil {
							conn = newConn
							// Resubscribe
							for _, pair := range pairs {
								subscribeMsg := map[string]interface{}{
									"event":   "subscribe",
									"channel": "trades",
									"symbols": []string{pair},
								}
								if err := conn.WriteJSON(subscribeMsg); err != nil {
									return
								}
							}
						}
					}
					continue
				}

				if msg.Event == "trade" {
					for _, trade := range msg.Data {
						select {
						case <-ctx.Done():
							return
						case trades <- models.RecentTrade{
							Tid:       trade.TradeID,
							Pair:      trade.Symbol,
							Price:     trade.Price,
							Amount:    trade.Quantity,
							Side:      trade.Side,
							Timestamp: trade.Ts,
						}:
						default:
							// Skip if channel is full
						}
					}
				}
			}
		}
	}()

	return trades, nil
}

func getTimeframeInterval(timeframe string) int64 {
	switch timeframe {
	case "1m":
		return 60
	case "15m":
		return 900
	case "1h":
		return 3600
	case "1d":
		return 86400
	default:
		return 60
	}
}
