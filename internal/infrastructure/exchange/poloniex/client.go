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
		startTime*1000,
		endTime*1000,
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

		// Since the API does not provide a division into nuy/sell
		// we can only roughly divide the volume in half
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

	// Используем пары в оригинальном формате BTC_USDT
	formattedPairs := make([]string, len(pairs))
	for i, pair := range pairs {
		formattedPairs[i] = strings.ToUpper(pair) // Убедимся, что пара в верхнем регистре
	}

	// Функция для создания подключения
	connectWebSocket := func() (*websocket.Conn, error) {
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
		return conn, nil
	}

	// Функция для подписки на пары
	subscribe := func(conn *websocket.Conn) error {
		subscribeMsg := map[string]interface{}{
			"event":   "subscribe",
			"channel": []string{"trades"},
			"symbols": formattedPairs,
		}

		msgBytes, _ := json.Marshal(subscribeMsg)
		log.Printf("Sending subscription message: %s", string(msgBytes))

		if err := conn.WriteJSON(subscribeMsg); err != nil {
			return fmt.Errorf("subscribe error: %w", err)
		}

		// Ждем подтверждения подписки
		var response map[string]interface{}
		if err := conn.ReadJSON(&response); err != nil {
			return fmt.Errorf("read subscription response error: %w", err)
		}
		log.Printf("Subscription response: %+v", response)

		return nil
	}

	go func() {
		defer close(trades)

		for {
			select {
			case <-ctx.Done():
				log.Println("Context cancelled, stopping WebSocket reader")
				return
			default:
				conn, err := connectWebSocket()
				if err != nil {
					log.Printf("Connection error: %v, retrying in 5 seconds...", err)
					time.Sleep(5 * time.Second)
					continue
				}

				// Установка таймаутов
				conn.SetReadDeadline(time.Now().Add(60 * time.Second))
				conn.SetPongHandler(func(string) error {
					conn.SetReadDeadline(time.Now().Add(60 * time.Second))
					return nil
				})

				// Подписка
				if err := subscribe(conn); err != nil {
					log.Printf("Subscription error: %v, retrying...", err)
					conn.Close()
					time.Sleep(time.Second)
					continue
				}

				// Пинги для поддержания соединения
				pingTicker := time.NewTicker(30 * time.Second)
				go func() {
					defer pingTicker.Stop()
					for {
						select {
						case <-ctx.Done():
							return
						case <-pingTicker.C:
							if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
								log.Printf("Ping error: %v", err)
								return
							}
						}
					}
				}()

			readLoop:
				for {
					select {
					case <-ctx.Done():
						conn.Close()
						return
					default:
						// Читаем сообщение
						_, message, err := conn.ReadMessage()
						if err != nil {
							if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
								log.Printf("WebSocket error: %v, reconnecting...", err)
							}
							conn.Close()
							break readLoop
						}

						// Логируем сырое сообщение для отладки
						log.Printf("Received raw message: %s", string(message))

						// Структура для распарсивания JSON
						var msg struct {
							Channel string `json:"channel"`
							Data    []struct {
								Symbol     string `json:"symbol"`
								Amount     string `json:"amount"`
								Quantity   string `json:"quantity"`
								TakerSide  string `json:"takerSide"`
								CreateTime int64  `json:"createTime"`
								Price      string `json:"price"`
								ID         string `json:"id"`
								Timestamp  int64  `json:"ts"`
							} `json:"data"`
						}

						// Декодируем JSON
						if err := json.Unmarshal(message, &msg); err != nil {
							log.Printf("Error parsing message: %v", err)
							continue
						}

						// Проверяем, что это сообщение о сделках
						if msg.Channel != "trades" {
							continue
						}
						log.Printf("Received %+v trades", msg)

						// Обрабатываем сделки
						for _, trade := range msg.Data {
							price, err := strconv.ParseFloat(trade.Price, 64)
							if err != nil {
								log.Printf("Error parsing price: %v", err)
								continue
							}
							amount, err := strconv.ParseFloat(trade.Amount, 64)
							if err != nil {
								log.Printf("Error parsing amount: %v", err)
								continue
							}
							quantity, err := strconv.ParseFloat(trade.Quantity, 64)
							if err != nil {
								log.Printf("Error parsing quantity: %v", err)
								continue
							}

							// Создаем объект сделки
							recentTrade := models.RecentTrade{
								Symbol:     trade.Symbol,
								Amount:     amount,
								Quantity:   quantity,
								TakerSide:  trade.TakerSide,
								Price:      price,
								CreateTime: trade.CreateTime,
								Timestamp:  trade.Timestamp,
								ID:         trade.ID,
							}

							// Отправляем в канал
							select {
							case trades <- recentTrade:
							default:
								log.Println("Trade channel is full, dropping trade")
							}
						}
					}
				}

				log.Println("Reconnecting after read loop end...")
				time.Sleep(time.Second)
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
