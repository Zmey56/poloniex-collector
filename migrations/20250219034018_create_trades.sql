-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS trades (
                        id BIGSERIAL PRIMARY KEY,
                        tid VARCHAR(255) NOT NULL,
                        pair VARCHAR(20) NOT NULL,
                        price VARCHAR(50) NOT NULL,
                        amount VARCHAR(50) NOT NULL,
                        side VARCHAR(4) NOT NULL,
                        timestamp BIGINT NOT NULL,
                        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                        UNIQUE(tid, pair)
);

CREATE INDEX idx_trades_pair_timestamp ON trades(pair, timestamp);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS trades;
-- +goose StatementEnd
