-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS trades (
                        id BIGSERIAL PRIMARY KEY,
                        tid VARCHAR(255) NOT NULL,
                        pair VARCHAR(20) NOT NULL,
                        price DECIMAL(20, 8) NOT NULL,
                        amount DECIMAL(20, 8) NOT NULL,
                        quantity DECIMAL(20, 8) NOT NULL,
                        side VARCHAR(4) NOT NULL,
                        timestamp BIGINT NOT NULL,
                        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                        UNIQUE(tid, pair)
);

CREATE INDEX idx_trades_pair_tid ON trades(pair, tid);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS trades;
-- +goose StatementEnd
