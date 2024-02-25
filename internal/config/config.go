package config

import (
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Configuration struct {
	Lxbd struct {
		Username string `validate:"required"`
		Password string `validate:"required"`
	}
	Jellyseerr struct {
		ApiKey        string   `mapstructure:"api_key" validate:"required"`
		BaseUrl       string   `mapstructure:"base_url" validate:"required"`
		RequestsLimit int      `mapstructure:"requests_limit"`
		Filters       []string `mapstructure:"filters"`
	}
	TMDb struct {
		ApiKey string `mapstructure:"api_key" validate:"required"`
	}
	Tasks struct {
		DLWatchlist string `mapstructure:"dl_watchlist"`
	}
}

var config Configuration

func GetConfig() *Configuration {
	return &config
}

func init() {
	viper.AddConfigPath("/config")
	viper.SetConfigName("lbxd_seerr")
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	viper.SetDefault("jellyseerr.requests_limit", -1)
	viper.SetDefault("tasks.dl_watchlist", "disabled")

	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	validate := validator.New()
	if err := validate.Struct(&config); err != nil {
		log.Fatalf("Missing required attributes %v\n", err)
	}
}
