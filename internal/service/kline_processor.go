package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/Zmey56/poloniex-collector/internal/domain/models"
)

type KlineRepository interface {
	SaveKline(ctx context.Context, kline models.Kline) error
	GetLastKline(ctx context.Context, pair, timeframe string) (*models.Kline, error)
	GetKlineByInterval(ctx context.Context, pair, timeframe string, beginTime int64) (*models.Kline, error)
}

type KlineProcessor struct {
	repository KlineRepository
}

func NewKlineProcessor(repository KlineRepository) *KlineProcessor {
	return &KlineProcessor{
		repository: repository,
	}
}

func (p *KlineProcessor) ProcessTrade(ctx context.Context, trade *models.RecentTrade) error {
	timeframes := []string{PoloniexTimeFrame1m, PoloniexTimeFrame15m, PoloniexTimeFrame1h, PoloniexTimeFrame1d}

	log.Printf("Processing trade: Pair=%s, Price=%s, Amount=%s, Side=%s, Timestamp=%d",
		trade.Pair, trade.Price, trade.Amount, trade.Side, trade.Timestamp)

	price, err := strconv.ParseFloat(trade.Price, 64)
	if err != nil {
		return fmt.Errorf("invalid price format: %w", err)
	}

	amount, err := strconv.ParseFloat(trade.Amount, 64)
	if err != nil {
		return fmt.Errorf("invalid amount format: %w", err)
	}

	quoteAmount := price * amount

	for _, timeframe := range timeframes {
		tfConvert := ConvertAPIToTimeFrame(timeframe)

		beginTime, endTime := getKlineTimestamps(trade.Timestamp, tfConvert)
		log.Printf("Kline timestamps: BeginTime=%d, EndTime=%d", beginTime, endTime)
		log.Printf("Converting timestamps: BeginTime=%s, EndTime=%s", time.Unix(0, beginTime), time.Unix(0, endTime))

		var klineExists bool
		// lastKline, err := p.repository.GetKlineByInterval(ctx, trade.Pair, timeframe, beginTime/1000000000)
		lastKline, err := p.repository.GetLastKline(ctx, trade.Pair, timeframe)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if errors.Is(err, sql.ErrNoRows) || lastKline == nil {
			log.Printf("No kline found Pair=%s, TimeFrame=%s, BeginTime=%d, creating new one", trade.Pair, tfConvert, beginTime)
			klineExists = false
		}

		if lastKline != nil && lastKline.UtcBegin >= beginTime {
			log.Printf("Found kline: Pair=%s, TimeFrame=%s, BeginTime=%d, EndTime=%d",
				lastKline.Pair, lastKline.TimeFrame, lastKline.UtcBegin, lastKline.UtcEnd)
			klineExists = true
		} else {
			log.Printf("No kline found Pair=%s, TimeFrame=%s, BeginTime=%d, creating new one", trade.Pair, tfConvert, beginTime)
			klineExists = false
		}

		log.Printf("GetLastKline result: Pair=%s, TimeFrame=%s, Error=%v, Found=%v, LastKline=%+v,  klineExists§=%v",
			trade.Pair, timeframe, err, lastKline != nil, lastKline, klineExists)

		if !klineExists {
			newKline := models.Kline{
				Pair:      trade.Pair,
				TimeFrame: timeframe,
				O:         price,
				H:         price,
				L:         price,
				C:         price,
				UtcBegin:  beginTime,
				UtcEnd:    endTime,
				BeginDt:   time.Unix(0, beginTime*(int64(time.Millisecond))).UTC(),
				EndDt:     time.Unix(0, endTime*(int64(time.Millisecond))).UTC(),
				VolumeBS: models.VBS{
					BuyBase:   0,
					SellBase:  0,
					BuyQuote:  0,
					SellQuote: 0,
				},
			}

			if trade.Side == "buy" {
				newKline.VolumeBS.BuyBase = amount
				newKline.VolumeBS.BuyQuote = quoteAmount
			} else {
				newKline.VolumeBS.SellBase = amount
				newKline.VolumeBS.SellQuote = quoteAmount
			}

			if err := p.repository.SaveKline(ctx, newKline); err != nil {
				return err
			}
			log.Printf("Creating new kline: Pair=%s, TimeFrame=%s, BeginTime=%d",
				newKline.Pair, newKline.TimeFrame, newKline.UtcBegin)
		} else {
			lastKline.H = math.Max(lastKline.H, price)
			lastKline.L = math.Min(lastKline.L, price)
			lastKline.C = price

			if trade.Side == "buy" {
				lastKline.VolumeBS.BuyBase += amount
				lastKline.VolumeBS.BuyQuote += quoteAmount
			} else {
				lastKline.VolumeBS.SellBase += amount
				lastKline.VolumeBS.SellQuote += quoteAmount
			}

			if err := p.repository.SaveKline(ctx, *lastKline); err != nil {
				return err
			}
			log.Printf("Updating kline: Pair=%s, TimeFrame=%s, BeginTime=%d",
				lastKline.Pair, lastKline.TimeFrame, lastKline.UtcBegin)
		}
	}
	return nil
}

func getKlineTimestamps(timestamp int64, timeFrame string) (int64, int64) {
	var t time.Time
	if timestamp > 1000000000000 {
		t = time.Unix(timestamp/1000, (timestamp%1000)*1000000).UTC()
	} else if timestamp > 1000000000 {
		t = time.Unix(timestamp, 0).UTC()
	} else {
		t = time.Unix(0, timestamp).UTC()
	}

	var beginTime time.Time
	var endTime time.Time

	switch timeFrame {
	case TimeFrame1m:
		beginTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)
		endTime = beginTime.Add(1 * time.Minute)
	case TimeFrame15m:
		minute := t.Minute() / 15 * 15
		beginTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), minute, 0, 0, time.UTC)
		endTime = beginTime.Add(15 * time.Minute)
	case TimeFrame1h:
		beginTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.UTC)
		endTime = beginTime.Add(1 * time.Hour)
	case TimeFrame1d:
		beginTime = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		endTime = beginTime.Add(24 * time.Hour)
	}

	return beginTime.Unix() * 1000, endTime.Unix() * 1000
}
