package models

type RecentTrade struct {
	Tid       string `json:"tid"`
	Pair      string `json:"pair"`
	Price     string `json:"price"`
	Amount    string `json:"amount"`
	Side      string `json:"side"`
	Timestamp int64  `json:"timestamp"`
}
