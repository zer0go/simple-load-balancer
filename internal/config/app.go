package config

import (
	"github.com/caarlos0/env/v11"
)

type App struct {
	Address                    string   `env:"ADDRESS" envDefault:"0.0.0.0:8080"`
	BackendUrls                []string `env:"BACKEND_URLS,required"`
	HealthCheckPath            string   `env:"HEALTH_CHECK_PATH" envDefault:""`
	HealthCheckIntervalSeconds int      `env:"HEALTH_CHECK_INTERVAL_SECONDS" envDefault:"5"`
}

var (
	config App
)

func Get() *App {
	return &config
}

func LoadAppConfig() error {
	if err := env.Parse(Get()); err != nil {
		return err
	}

	return nil
}
