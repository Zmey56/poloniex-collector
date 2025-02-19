package models

type Kline struct {
	Pair      string  `json:"pair"`
	TimeFrame string  `json:"timeFrame"`
	O         float64 `json:"o"`
	H         float64 `json:"h"`
	L         float64 `json:"l"`
	C         float64 `json:"c"`
	UtcBegin  int64   `json:"utcBegin"`
	UtcEnd    int64   `json:"utcEnd"`
	VolumeBS  VBS     `json:"volumeBS"`
}
