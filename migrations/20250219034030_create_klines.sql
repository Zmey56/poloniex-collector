-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS klines (
                        id SERIAL PRIMARY KEY,
                        pair VARCHAR(20) NOT NULL,
                        interval VARCHAR(10) NOT NULL,
                        open DECIMAL(20, 8) NOT NULL,
                        high DECIMAL(20, 8) NOT NULL,
                        low DECIMAL(20, 8) NOT NULL,
                        close DECIMAL(20, 8) NOT NULL,
                        utc_begin BIGINT NOT NULL,
                        utc_end BIGINT NOT NULL,
                        begin_dt TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        end_dt TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        volume_bs JSONB NOT NULL,
                        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                        UNIQUE(pair, interval, utc_begin)
);

CREATE INDEX idx_klines_pair_timeframe_utc ON klines(pair, interval, utc_begin);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS klines;
-- +goose StatementEnd
