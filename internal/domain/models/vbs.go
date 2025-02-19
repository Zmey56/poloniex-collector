package models

type VBS struct {
	BuyBase   float64 `json:"buyBase"`
	SellBase  float64 `json:"sellBase"`
	BuyQuote  float64 `json:"buyQuote"`
	SellQuote float64 `json:"sellQuote"`
}
