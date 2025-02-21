package config

import (
	"errors"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Name     string `mapstructure:"name"`
		SSLMode  string `mapstructure:"sslmode"`
	} `mapstructure:"database"`

	Poloniex struct {
		WSURL      string   `mapstructure:"ws_url"`
		RestURL    string   `mapstructure:"rest_url"`
		Pairs      []string `mapstructure:"pairs"`
		TimeFrames []string `mapstructure:"timeframes"`
	} `mapstructure:"poloniex"`

	Worker struct {
		PoolSize      int           `mapstructure:"pool_size"`
		BatchSize     int           `mapstructure:"batch_size"`
		FlushInterval time.Duration `mapstructure:"flush_interval"`
	} `mapstructure:"worker"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Default values
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.name", "poloniex")
	viper.SetDefault("database.sslmode", "disable")

	viper.SetDefault("poloniex.ws_url", "wss://ws.poloniex.com/ws/public")
	viper.SetDefault("poloniex.rest_url", "https://api.poloniex.com")
	viper.SetDefault("poloniex.pairs", []string{"BTC_USDT", "ETH_USDT", "TRX_USDT", "DOGE_USDT", "BCH_USDT"})
	viper.SetDefault("poloniex.timeframes", []string{"MINUTE_1", "MINUTE_15", "HOUR_1", "DAY_1"})

	viper.SetDefault("worker.pool_size", 10)
	viper.SetDefault("worker.batch_size", 1000)
	viper.SetDefault("worker.flush_interval", "5s")

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
