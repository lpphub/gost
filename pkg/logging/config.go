package logging

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

// 定义别名（抽象提取，可替换不同实现）
type (
	logger = zerolog.Logger
	Event  = zerolog.Event
	Level  = zerolog.Level
)

func newZerolog(cfg *config) logger {
	// 全局配置
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000Z07:00"
	zerolog.CallerMarshalFunc = callerShortFunc

	base := zerolog.New(cfg.output).
		Level(cfg.level).
		With().
		Timestamp().
		Logger()
	return base
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

func WithOutputFile(filepath string) Option {
	return func(c *config) {
		lj := &lumberjack.Logger{
			Filename:   filepath, // 路径
			MaxSize:    200,      // 单个文件最大 MB
			MaxBackups: 5,        // 最多保留几个旧文件
			MaxAge:     14,       // 旧文件最长保存天数
			Compress:   true,     // 是否 gzip 压缩旧文件
		}
		bufWriter := bufio.NewWriter(lj)
		c.output = bufWriter
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
