package poloniex

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
	log.Printf("Getting historical klines for pair: %s, timeframe: %s", pair, timeframe)
	u := fmt.Sprintf("%s/markets/%s/candles?interval=%s&startTime=%d&endTime=%d",
		c.restURL,
		pair,
		timeframe,
		startTime*1000, // Convert to milliseconds
		endTime*1000,   // Convert to milliseconds
	)

	log.Printf("Making request to: %s", u)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	// Decode JSON as an array of arrays
	var rawData [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	log.Printf("Received %d klines from API", len(rawData))

	klines := make([]models.Kline, len(rawData))
	for i, row := range rawData {
		if len(row) < 14 {
			log.Printf("Skipping malformed entry: %v", row)
			continue
		}

		// Convert values from interface{} to appropriate types
		open, _ := strconv.ParseFloat(row[0].(string), 64)
		high, _ := strconv.ParseFloat(row[1].(string), 64)
		low, _ := strconv.ParseFloat(row[2].(string), 64)
		close, _ := strconv.ParseFloat(row[3].(string), 64)
		volume, _ := strconv.ParseFloat(row[4].(string), 64)
		startTimestamp := int64(row[12].(float64)) / 1000 // Convert ms to sec
		endTimestamp := int64(row[13].(float64)) / 1000   // Convert ms to sec

		klines[i] = models.Kline{
			Pair:      pair,
			TimeFrame: timeframe,
			O:         open,
			H:         high,
			L:         low,
			C:         close,
			UtcBegin:  startTimestamp,
			UtcEnd:    endTimestamp,
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
	log.Printf("Starting subscription to trades for pairs: %v", pairs)
	trades := make(chan models.RecentTrade, 1000)

	u, err := url.Parse(c.wsURL)
	if err != nil {
		return nil, fmt.Errorf("parse ws url error: %w", err)
	}
	log.Printf("Connecting to WebSocket: %s", u.String())

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("dial ws error: %w", err)
	}
	log.Println("WebSocket connection established")

	// Подписываемся на каналы
	for _, pair := range pairs {
		// Заменяем "_" на "/" для формата Poloniex
		formattedPair := strings.Replace(pair, "_", "/", -1)
		subscribeMsg := map[string]interface{}{
			"event":   "subscribe",
			"channel": []string{"trades"},
			"symbols": []string{strings.Replace(pair, "_", "/", -1)},
		}

		log.Printf("Sending subscription message for pair: %s", formattedPair)
		msgBytes, _ := json.Marshal(subscribeMsg)
		log.Printf("Raw subscription message: %s", string(msgBytes))

		if err := conn.WriteJSON(subscribeMsg); err != nil {
			conn.Close()
			return nil, fmt.Errorf("subscribe error for pair %s: %w", pair, err)
		}
		log.Printf("Successfully subscribed to pair: %s", formattedPair)
	}

	go func() {
		defer func() {
			log.Println("Closing WebSocket connection")
			conn.Close()
			close(trades)
		}()

		for {
			select {
			case <-ctx.Done():
				log.Println("Context cancelled, stopping WebSocket reader")
				return
			default:
				// Читаем сообщение
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Printf("Error reading WebSocket message: %v", err)
					if websocket.IsUnexpectedCloseError(err) {
						log.Printf("WebSocket closed unexpectedly: %v", err)
						// Пытаемся переподключиться
						time.Sleep(time.Second)
						if newConn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil); err == nil {
							conn = newConn
							log.Println("Successfully reconnected to WebSocket")
							// Повторная подписка
							for _, pair := range pairs {
								formattedPair := strings.Replace(pair, "_", "/", -1)
								subscribeMsg := map[string]interface{}{
									"event":   "subscribe",
									"channel": []string{"trades"},
									"symbols": []string{formattedPair},
								}
								if err := conn.WriteJSON(subscribeMsg); err != nil {
									log.Printf("Failed to resubscribe to pair %s: %v", pair, err)
									return
								}
								log.Printf("Successfully resubscribed to pair: %s", formattedPair)
							}
						}
					}
					continue
				}

				// Логируем raw сообщение
				log.Printf("Received raw message: %s", string(message))

				// Пытаемся распарсить сообщение
				var msg struct {
					Event   string `json:"event"`
					Channel string `json:"channel"`
					Data    []struct {
						Symbol   string `json:"symbol"`
						TradeID  string `json:"id"`
						Price    string `json:"price"`
						Quantity string `json:"amount"`
						Side     string `json:"type"`
						Ts       int64  `json:"timestamp"`
					} `json:"data"`
				}

				if err := json.Unmarshal(message, &msg); err != nil {
					log.Printf("Error unmarshaling message: %v", err)
					continue
				}

				log.Printf("Parsed message: event=%s, channel=%s, data_length=%d",
					msg.Event, msg.Channel, len(msg.Data))

				if msg.Event == "trade" {
					for _, trade := range msg.Data {
						log.Printf("Processing trade: ID=%s, Symbol=%s, Price=%s, Amount=%s",
							trade.TradeID, trade.Symbol, trade.Price, trade.Quantity)

						// Заменяем "/" обратно на "_" для соответствия нашему формату
						pair := strings.Replace(trade.Symbol, "/", "_", -1)

						select {
						case trades <- models.RecentTrade{
							Tid:       trade.TradeID,
							Pair:      pair,
							Price:     trade.Price,
							Amount:    trade.Quantity,
							Side:      trade.Side,
							Timestamp: trade.Ts,
						}:
							log.Printf("Successfully sent trade to channel: %s", trade.TradeID)
						default:
							log.Println("Warning: trade channel is full, skipping trade")
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
