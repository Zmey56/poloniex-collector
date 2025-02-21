package postgres

import (
	"context"
	"encoding/json"

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
	volumeBS, err := json.Marshal(kline.VolumeBS)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO klines (
            pair, interval, open, high, low, close, 
            utc_begin, utc_end, volume_bs, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
        ON CONFLICT (pair, interval, utc_begin) DO UPDATE
        SET 
            high = GREATEST(klines.high, EXCLUDED.high),
            low = LEAST(klines.low, EXCLUDED.low),
            close = EXCLUDED.close,
            volume_bs = EXCLUDED.volume_bs,
            updated_at = NOW()`,
		kline.Pair,
		kline.TimeFrame,
		kline.O,
		kline.H,
		kline.L,
		kline.C,
		kline.UtcBegin,
		kline.UtcEnd,
		volumeBS)

	return err
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
