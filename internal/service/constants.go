package service

const (
	TimeFrame1m  = "1m"
	TimeFrame15m = "15m"
	TimeFrame1h  = "1h"
	TimeFrame1d  = "1d"

	PoloniexTimeFrame1m  = "MINUTE_1"
	PoloniexTimeFrame15m = "MINUTE_15"
	PoloniexTimeFrame1h  = "HOUR_1"
	PoloniexTimeFrame1d  = "DAY_1"
)

var TimeFrameMapping = map[string]string{
	TimeFrame1m:  PoloniexTimeFrame1m,
	TimeFrame15m: PoloniexTimeFrame15m,
	TimeFrame1h:  PoloniexTimeFrame1h,
	TimeFrame1d:  PoloniexTimeFrame1d,
}

func ConvertTimeFrameToAPI(timeframe string) string {
	if apiTimeframe, ok := TimeFrameMapping[timeframe]; ok {
		return apiTimeframe
	}
	return timeframe
}

func ConvertAPIToTimeFrame(apiTimeframe string) string {
	for internal, api := range TimeFrameMapping {
		if api == apiTimeframe {
			return internal
		}
	}
	return apiTimeframe
}

func GetTimeFrameDuration(timeframe string) int64 {
	switch timeframe {
	case TimeFrame1m, PoloniexTimeFrame1m:
		return 60 * 1e9
	case TimeFrame15m, PoloniexTimeFrame15m:
		return 15 * 60 * 1e9
	case TimeFrame1h, PoloniexTimeFrame1h:
		return 60 * 60 * 1e9
	case TimeFrame1d, PoloniexTimeFrame1d:
		return 24 * 60 * 60 * 1e9
	default:
		return 60 * 1e9
	}
}
