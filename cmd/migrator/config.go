package main

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type config struct {
	PG pgConfig
}

type pgConfig struct {
	Addr     string `env:"PG_ADDRESS" env-description:"Postgres address (host:port)" env-default:"localhost:6432"`
	Database string `env:"PG_DATABASE" env-description:"Postgres database name" env-default:"simple-storage"`
	User     string `env:"PG_USER" env-description:"Postgres username" env-default:"simple-storage-user"`
	Password string `env:"PG_PASSWORD" env-description:"Postgres password"`
}

func newConfig() (config, error) {
	var cfg config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	return cfg, nil
}
