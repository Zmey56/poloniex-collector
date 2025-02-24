package config

import (
	"errors"
	"log"
	"os"
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

	log.Printf("Environment variables:")
	log.Printf("DATABASE_HOST=%s", os.Getenv("DATABASE_HOST"))
	log.Printf("DATABASE_PORT=%s", os.Getenv("DATABASE_PORT"))
	log.Printf("DATABASE_USER=%s", os.Getenv("DATABASE_USER"))
	log.Printf("DATABASE_NAME=%s", os.Getenv("DATABASE_NAME"))
	log.Printf("DATABASE_SSLMODE=%s", os.Getenv("DATABASE_SSLMODE"))

	err := viper.BindEnv("database.host", "DATABASE_HOST")
	if err != nil {
		log.Println("Failed to bind environment variable DATABASE_HOST")
		return nil, err
	}
	err = viper.BindEnv("database.port", "DATABASE_PORT")
	if err != nil {
		log.Println("Failed to bind environment variable DATABASE_PORT")
		return nil, err
	}
	err = viper.BindEnv("database.user", "DATABASE_USER")
	if err != nil {
		log.Println("Failed to bind environment variable DATABASE_USER")
		return nil, err
	}
	err = viper.BindEnv("database.password", "DATABASE_PASSWORD")
	if err != nil {
		log.Println("Failed to bind environment variable DATABASE_PASSWORD")
		return nil, err
	}
	err = viper.BindEnv("database.name", "DATABASE_NAME")
	if err != nil {
		log.Println("Failed to bind environment variable DATABASE_NAME")
		return nil, err
	}
	err = viper.BindEnv("database.sslmode", "DATABASE_SSLMODE")
	if err != nil {
		log.Println("Failed to bind environment variable DATABASE_SSLMODE")
		return nil, err
	}

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

	log.Printf("Final configuration:")
	log.Printf("database.host=%s", viper.GetString("database.host"))
	log.Printf("database.port=%d", viper.GetInt("database.port"))
	log.Printf("database.user=%s", viper.GetString("database.user"))
	log.Printf("database.name=%s", viper.GetString("database.name"))
	log.Printf("database.sslmode=%s", viper.GetString("database.sslmode"))

	return &config, nil
}
