package logging

import (
	"time"
)

type Field func(e *Event)

// Str 返回字符串字段
func Str(key, val string) Field {
	return func(e *Event) { e.Str(key, val) }
}

// Int 返回整数字段
func Int(key string, val int) Field {
	return func(e *Event) { e.Int(key, val) }
}

// Int64 返回 int64 字段
func Int64(key string, val int64) Field {
	return func(e *Event) { e.Int64(key, val) }
}

// Dur 返回 duration 字段
func Dur(key string, val time.Duration) Field {
	return func(e *Event) { e.Dur(key, val) }
}

// Err 返回错误字段
func Err(err error) Field {
	return func(e *Event) {
		if err != nil {
			e.Err(err)
		}
	}
}

// Caller 返回 caller 字段
func Caller(skip int) Field {
	return func(e *Event) { e.Caller(skip) }
}
