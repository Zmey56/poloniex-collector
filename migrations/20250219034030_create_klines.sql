-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS klines (
                        id SERIAL PRIMARY KEY,
                        pair VARCHAR(20) NOT NULL,
                        timeframe VARCHAR(10) NOT NULL,
                        open DECIMAL(20, 8) NOT NULL,
                        high DECIMAL(20, 8) NOT NULL,
                        low DECIMAL(20, 8) NOT NULL,
                        close DECIMAL(20, 8) NOT NULL,
                        utc_begin BIGINT NOT NULL,
                        utc_end BIGINT NOT NULL,
                        volume_bs JSONB NOT NULL,
                        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                        UNIQUE(pair, timeframe, utc_begin)
);

CREATE INDEX idx_klines_pair_timeframe_utc ON klines(pair, timeframe, utc_begin);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS klines;
-- +goose StatementEnd
