package log

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type (
	Event  = zerolog.Event
	Level  = zerolog.Level
)

var timeFormat = "2006-01-02 15:04:05.000Z07:00"

func newZerolog(cfg *config) zerolog.Logger {
	zerolog.TimeFieldFormat = timeFormat
	zerolog.CallerMarshalFunc = callerShortFunc

	return zerolog.New(cfg.output).
		Level(cfg.level).
		With().
		Timestamp().
		Logger()
}

type config struct {
	level  Level
	output io.Writer
}

type Option func(*config)

func defaultConfig() *config {
	return &config{
		level:  zerolog.InfoLevel,
		output: os.Stdout,
	}
}

func WithLevel(level Level) Option {
	return func(c *config) {
		c.level = level
	}
}

func WithOutput(w io.Writer) Option {
	return func(c *config) {
		c.output = w
	}
}

func WithOutputFile(path string) Option {
	return func(c *config) {
		lj := &lumberjack.Logger{
			Filename:   path,
			MaxSize:    200,
			MaxBackups: 5,
			MaxAge:     14,
			Compress:   true,
		}
		c.output = bufio.NewWriter(lj)
	}
}

func callerShortFunc(_ uintptr, file string, line int) string {
	file = filepath.ToSlash(file)
	parts := strings.Split(file, "/")
	if len(parts) > 2 {
		file = strings.Join(parts[len(parts)-2:], "/")
	}
	return fmt.Sprintf("%s:%d", file, line)
}
