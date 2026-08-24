package log

import (
	"github.com/rs/zerolog"
)

var std *zerolog.Logger

var nopLogger = zerolog.Nop()

func Init(opts ...Option) error {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}
	if cfg.err != nil {
		return cfg.err
	}

	l := newZerolog(cfg)
	std = &l
	zerolog.DefaultContextLogger = std
	return nil
}

func L() *zerolog.Logger {
	if std == nil {
		return &nopLogger
	}
	return std
}
