package main

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type config struct {
	Addr        string `env:"GRPC_LISTEN_ADDR" env-default:":50051"`
	StoragePath string `env:"STORAGE_PATH" env-default:"/storage"`
}

func newConfig() (config, error) {
	var cfg config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	return cfg, nil
}
