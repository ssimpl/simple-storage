package main

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type config struct {
	Addr              string        `env:"HTTP_LISTEN_ADDR" env-default:":8080"`
	ConnectionTimeout time.Duration `env:"CONNECTION_TIMEOUT" env-default:"5s"`
	FileFragments     int           `env:"FILE_FRAGMENTS" env-default:"6"`
	FileSizeLimit     int64         `env:"FILE_SIZE_LIMIT" env-default:"10737418240" env-description:"Default: 10 GB"`

	PG pgConfig
}

type pgConfig struct {
	Addr     string        `env:"PG_ADDRESS" env-description:"Postgres address (host:port)" env-default:"localhost:6432"`
	Database string        `env:"PG_DATABASE" env-description:"Postgres database name" env-default:"simple-storage"`
	User     string        `env:"PG_USER" env-description:"Postgres username" env-default:"simple-storage-user"`
	Password string        `env:"PG_PASSWORD" env-description:"Postgres password"`
	Timeout  time.Duration `env:"PG_TIMEOUT" env-description:"Query timeout" env-default:"10s"`
	AppName  string        `env:"PG_APP_NAME" env-description:"Log queries from app name" env-default:"simple-storage-api"`
	SQLDebug bool          `env:"PG_SQL_DEBUG" env-description:"Toggle SQL debug"`
}

func newConfig() (config, error) {
	var cfg config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	return cfg, nil
}
