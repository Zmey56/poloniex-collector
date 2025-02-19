package postgres

import (
	"context"
	"github.com/Zmey56/poloniex-collector/internal/domain/models"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type TradeRepository struct {
	pool *pgxpool.Pool
}

func NewTradeRepository(pool *pgxpool.Pool) *TradeRepository {
	return &TradeRepository{
		pool: pool,
	}
}

func (r *TradeRepository) SaveTrade(ctx context.Context, trade models.RecentTrade) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO trades (tid, pair, price, amount, side, timestamp)
         VALUES ($1, $2, $3, $4, $5, $6)
         ON CONFLICT (tid, pair) DO NOTHING`,
		trade.Tid, trade.Pair, trade.Price, trade.Amount, trade.Side, trade.Timestamp)
	return err
}

func (r *TradeRepository) SaveTrades(ctx context.Context, trades []models.RecentTrade) error {
	batch := &pgx.Batch{}

	for _, trade := range trades {
		batch.Queue(
			`INSERT INTO trades (tid, pair, price, amount, side, timestamp)
             VALUES ($1, $2, $3, $4, $5, $6)
             ON CONFLICT (tid, pair) DO NOTHING`,
			trade.Tid, trade.Pair, trade.Price, trade.Amount, trade.Side, trade.Timestamp)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	return nil
}
