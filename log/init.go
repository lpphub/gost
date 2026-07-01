package log

import (
	"os"

	"github.com/rs/zerolog"
)

var std = func() *zerolog.Logger {
	l := zerolog.New(os.Stdout).With().Timestamp().Logger()
	return &l
}()

func Init(opts ...Option) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}
	l := newZerolog(cfg)
	std = &l
}

func L() *zerolog.Logger {
	return std
}
