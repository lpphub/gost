package log

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type (
	Event = zerolog.Event
	Level = zerolog.Level
)

var (
	timeFormat = "2006-01-02 15:04:05.000Z07:00"
	zlogOnce   sync.Once
)

func newZerolog(cfg *config) zerolog.Logger {
	zlogOnce.Do(func() {
		zerolog.TimeFieldFormat = timeFormat
		zerolog.CallerMarshalFunc = callerShortFunc
	})

	return zerolog.New(cfg.writer()).
		Level(cfg.level).
		With().
		Timestamp().
		Caller().
		Logger()
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

func WithWriter(w io.Writer) Option {
	return func(c *config) {
		c.outputs = append(c.outputs, w)
	}
}

func WithFileWriter(path string) Option {
	return func(c *config) {
		lj := &lumberjack.Logger{
			Filename:   path,
			MaxSize:    200,
			MaxBackups: 5,
			MaxAge:     14,
			Compress:   true,
		}
		c.outputs = append(c.outputs, bufio.NewWriter(lj))
	}
}

func (c *config) writer() io.Writer {
	if len(c.outputs) == 0 {
		return os.Stdout
	}
	if len(c.outputs) == 1 {
		return c.outputs[0]
	}
	return io.MultiWriter(c.outputs...)
}

func callerShortFunc(_ uintptr, file string, line int) string {
	file = filepath.ToSlash(file)
	n := 2
	for i := len(file) - 1; i >= 0; i-- {
		if file[i] == '/' {
			n--
			if n == 0 {
				file = file[i+1:]
				break
			}
		}
	}
	return file + ":" + strconv.Itoa(line)
}
