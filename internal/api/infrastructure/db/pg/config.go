package pg

import (
	"time"
)

const (
	defaultTimeout = 10 * time.Second
)

type Config struct {
	Addr     string
	Database string
	User     string
	Password string
	SQLDebug bool
	AppName  string
	Timeout  time.Duration
}

func (cfg *Config) SetDefaults() {
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
}
