package log

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	glog "github.com/lpphub/gost/log"
	"github.com/redis/go-redis/extra/rediscmd/v9"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type RedisLogCfg struct {
	SlowThreshold time.Duration
	CmdMaxLen     int
	CallerSkip    int
	Level         zerolog.Level
}

type RedisLogger struct {
	cfg RedisLogCfg
}

func NewRedisLogger(cfg RedisLogCfg) *RedisLogger {
	if cfg.SlowThreshold == 0 {
		cfg.SlowThreshold = 100 * time.Millisecond
	}
	if cfg.CmdMaxLen == 0 {
		cfg.CmdMaxLen = 1024
	}
	return &RedisLogger{cfg: cfg}
}

func (l *RedisLogger) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		start := time.Now()
		conn, err := next(ctx, network, addr)
		cost := time.Since(start).Milliseconds()

		if err != nil {
			if ev := l.event(ctx, zerolog.ErrorLevel); ev != nil {
				ev.Err(err).Str("addr", addr).Int64("cost", cost).Msg("redis connected failed")
			}
		} else {
			if ev := l.event(ctx, zerolog.InfoLevel); ev != nil {
				ev.Str("addr", addr).Int64("cost", cost).Msg("redis connected")
			}
		}
		return conn, err
	}
}

func (l *RedisLogger) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)
		elapsed := time.Since(start)

		l.logResult(ctx, err, elapsed, l.buildCmd(cmd), "redis success", "redis slow", "redis error")
		return err
	}
}

func (l *RedisLogger) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmds)
		elapsed := time.Since(start)

		l.logResult(ctx, err, elapsed, l.buildPipelineCmd(cmds), "redis pipeline success", "redis pipeline slow", "redis pipeline error")
		return err
	}
}

func (l *RedisLogger) logResult(ctx context.Context, err error, elapsed time.Duration, cmd, okMsg, slowMsg, errMsg string) {
	switch {
	case err != nil && !errors.Is(err, redis.Nil):
		if ev := l.event(ctx, zerolog.ErrorLevel); ev != nil {
			ev.Err(err).Str("cmd", cmd).Int64("cost", elapsed.Milliseconds()).Msg(errMsg)
		}
	case l.cfg.SlowThreshold > 0 && elapsed > l.cfg.SlowThreshold:
		if ev := l.event(ctx, zerolog.WarnLevel); ev != nil {
			ev.Str("cmd", cmd).Int64("cost", elapsed.Milliseconds()).Msg(slowMsg)
		}
	default:
		if ev := l.event(ctx, zerolog.InfoLevel); ev != nil {
			ev.Str("cmd", cmd).Int64("cost", elapsed.Milliseconds()).Msg(okMsg)
		}
	}
}

func (l *RedisLogger) event(ctx context.Context, level zerolog.Level) *zerolog.Event {
	if l.cfg.Level != 0 && level < l.cfg.Level {
		return nil
	}
	var ev *zerolog.Event
	switch level {
	case zerolog.InfoLevel:
		ev = glog.Ctx(ctx).Info()
	case zerolog.WarnLevel:
		ev = glog.Ctx(ctx).Warn()
	case zerolog.ErrorLevel:
		ev = glog.Ctx(ctx).Error()
	default:
		return nil
	}
	return withCaller(ev, l.cfg.CallerSkip)
}

func (l *RedisLogger) buildCmd(cmd redis.Cmder) string {
	return truncate(rediscmd.CmdString(cmd), l.cfg.CmdMaxLen)
}

func (l *RedisLogger) buildPipelineCmd(cmds []redis.Cmder) string {
	if len(cmds) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, cmd := range cmds {
		if i > 0 {
			sb.WriteString("; ")
		}
		if i >= 5 {
			sb.WriteString("... (")
			sb.WriteString(strconv.Itoa(len(cmds) - 5))
			sb.WriteString(" more)")
			break
		}
		sb.WriteString(rediscmd.CmdString(cmd))
	}
	return truncate(sb.String(), l.cfg.CmdMaxLen)
}
