package models

type RecentTrade struct {
	Tid       string `json:"id"`
	Pair      string `json:"pair"`
	Price     string `json:"price"`
	Amount    string `json:"amount"`
	Side      string `json:"taker_side"`
	Timestamp int64  `json:"timestamp"`

	Symbol     string  `json:"symbol"`
	Quantity   float64 `json:"quantity"`
	CreateTime int64   `json:"create_time"`
}
