package main

import "github.com/caarlos0/env/v6"

type config struct {
	Listen   string `env:"LISTEN" envDefault:"localhost:7171"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"debug"`
	LogFmt   string `env:"LOG_FMT" envDefault:"console"`
	DbAddr string `env:"DB_HOST" envDefault:"postgresql://postgres:password@localhost:5432/zoodb?sslmode=disable&application_name=zooadmin"`
}

func initConfig() (*config, error) {
	cfg := &config{}

	if err := env.Parse(cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
