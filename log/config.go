package log

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
)

type (
	Event = zerolog.Event
	Level = zerolog.Level
)

func newZerolog(cfg *config) zerolog.Logger {
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000Z07:00"
	zerolog.CallerMarshalFunc = callerShortFunc

	return zerolog.New(cfg.writer()).
		Level(cfg.level).With().Timestamp().Logger()
}

type config struct {
	level   Level
	outputs []io.Writer
}

type Option func(*config)

func defaultConfig() *config {
	return &config{
		level:   zerolog.InfoLevel,
		outputs: []io.Writer{os.Stdout},
	}
}

func WithLevel(level Level) Option {
	return func(c *config) {
		c.level = level
	}
}

// WithWriter adds a log output sink. Pass it multiple times to write to
// several sinks at once (they are combined with io.MultiWriter).
func WithWriter(w io.Writer) Option {
	return func(c *config) {
		c.outputs = append(c.outputs, w)
	}
}

func (c *config) writer() io.Writer {
	if len(c.outputs) == 1 {
		return c.outputs[0]
	}
	return io.MultiWriter(c.outputs...)
}

func callerShortFunc(_ uintptr, file string, line int) string {
	parts := strings.Split(filepath.ToSlash(file), "/")
	if len(parts) > 2 {
		parts = parts[len(parts)-2:]
	}
	return strings.Join(parts, "/") + ":" + strconv.Itoa(line)
}
