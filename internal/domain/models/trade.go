package models

type RecentTrade struct {
	Symbol     string  `json:"symbol"`
	Tid        int64   `json:"tid"`
	Pair       string  `json:"pair"`
	Price      float64 `json:"price"`
	Amount     float64 `json:"amount"`
	Side       string  `json:"side"`
	Timestamp  int64   `json:"timestamp"`
	Quantity   float64 `json:"quantity"`
	TakerSide  string  `json:"taker_side"`
	CreateTime int64   `json:"create_time"`
	ID         string  `json:"id"`
}
