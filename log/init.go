package log

import (
	"github.com/rs/zerolog"
)

var std *zerolog.Logger

func Init(opts ...Option) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}
	l := newZerolog(cfg)
	std = &l
	zerolog.DefaultContextLogger = std
}

func L() *zerolog.Logger {
	if std == nil {
		l := zerolog.Nop()
		return &l
	}
	return std
}
