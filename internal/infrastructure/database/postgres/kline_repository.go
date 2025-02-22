package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/Zmey56/poloniex-collector/internal/domain/models"
)

type KlineRepository struct {
	pool *pgxpool.Pool
}

func NewKlineRepository(pool *pgxpool.Pool) *KlineRepository {
	return &KlineRepository{
		pool: pool,
	}
}

func (r *KlineRepository) SaveKline(ctx context.Context, kline models.Kline) error {
	log.Printf("Saving kline in repository: Pair=%s, Timeframe=%s, UtcBegin=%d", kline.Pair, kline.TimeFrame, kline.UtcBegin)
	volumeBSJson, err := json.Marshal(kline.VolumeBS)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO klines (pair, interval, open, high, low, close, utc_begin, utc_end, volume_bs, begin_dt, end_dt)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
         ON CONFLICT (pair, interval, utc_begin) 
         DO UPDATE SET
            high = GREATEST(klines.high, $4),
            low = LEAST(klines.low, $5),
            close = $6,
            volume_bs = $9`,
		kline.Pair, kline.TimeFrame, kline.O, kline.H, kline.L, kline.C,
		kline.UtcBegin, kline.UtcEnd, volumeBSJson, kline.BeginDt, kline.EndDt)

	return err
}

func (r *KlineRepository) GetKlineByInterval(ctx context.Context, pair, timeframe string, beginTime int64) (*models.Kline, error) {
	var kline models.Kline
	log.Printf("Getting kline by interval in repository: Pair=%s, Timeframe=%s, BeginTime=%d", pair, timeframe, beginTime)

	query := `SELECT id, pair, "interval", "open", high, low, "close", utc_begin, utc_end, volume_bs, created_at, updated_at
              FROM klines 
              WHERE pair = $1 AND interval = $2 AND utc_begin >= $3`

	err := r.pool.QueryRow(ctx, query, pair, timeframe, beginTime).Scan(
		&kline.Pair, &kline.TimeFrame, &kline.O, &kline.H, &kline.L, &kline.C,
		&kline.UtcBegin, &kline.UtcEnd,
		&kline.VolumeBS.BuyBase, &kline.VolumeBS.SellBase,
		&kline.VolumeBS.BuyQuote, &kline.VolumeBS.SellQuote,
	)

	if errors.Is(err, sql.ErrNoRows) {
		log.Println("No rows in result set")
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &kline, nil
}

// count in table
func (r *KlineRepository) CountKlines(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM klines").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *KlineRepository) GetLastKline(ctx context.Context, pair, timeframe string) (*models.Kline, error) {
	var kline models.Kline
	var volumeBSJson []byte
	log.Printf("Getting last kline in repository: Pair=%s, Timeframe=%s", pair, timeframe)

	err := r.pool.QueryRow(ctx,
		`SELECT pair, interval, open, high, low, close, 
                utc_begin, utc_end, volume_bs
         FROM klines
         WHERE pair = $1 AND interval = $2
         ORDER BY utc_begin DESC
         LIMIT 1`,
		pair, timeframe).Scan(
		&kline.Pair,
		&kline.TimeFrame,
		&kline.O,
		&kline.H,
		&kline.L,
		&kline.C,
		&kline.UtcBegin,
		&kline.UtcEnd,
		&volumeBSJson)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(volumeBSJson, &kline.VolumeBS); err != nil {
		return nil, err
	}

	return &kline, nil
}

func (r *KlineRepository) GetKlinesByTimeRange(ctx context.Context, pair, timeframe string, startTime, endTime int64) ([]models.Kline, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT pair, interval, open, high, low, close, 
                utc_begin, utc_end, volume_bs
         FROM klines
         WHERE pair = $1 
           AND interval = $2 
           AND utc_begin >= $3 
           AND utc_end <= $4
         ORDER BY utc_begin`,
		pair, timeframe, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var klines []models.Kline
	for rows.Next() {
		var kline models.Kline
		var volumeBSJson []byte

		if err := rows.Scan(
			&kline.Pair,
			&kline.TimeFrame,
			&kline.O,
			&kline.H,
			&kline.L,
			&kline.C,
			&kline.UtcBegin,
			&kline.UtcEnd,
			&volumeBSJson); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(volumeBSJson, &kline.VolumeBS); err != nil {
			return nil, err
		}

		klines = append(klines, kline)
	}

	return klines, nil
}
