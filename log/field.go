package log

import (
	"time"

	"github.com/rs/zerolog"
)

type fieldKind int

const (
	fieldStr fieldKind = iota
	fieldInt
	fieldInt64
	fieldFloat64
	fieldBool
	fieldDur
	fieldErr
	fieldAny
)

type Field struct {
	kind     fieldKind
	key      string
	strVal   string
	intVal   int
	int64Val int64
	floatVal float64
	boolVal  bool
	durVal   time.Duration
	errVal   error
	anyVal   any
}

func (f Field) ApplyCtx(c zerolog.Context) zerolog.Context {
	switch f.kind {
	case fieldStr:
		return c.Str(f.key, f.strVal)
	case fieldInt:
		return c.Int(f.key, f.intVal)
	case fieldInt64:
		return c.Int64(f.key, f.int64Val)
	case fieldFloat64:
		return c.Float64(f.key, f.floatVal)
	case fieldBool:
		return c.Bool(f.key, f.boolVal)
	case fieldDur:
		return c.Dur(f.key, f.durVal)
	default:
		return c
	}
}

func (f Field) ApplyEvent(e *zerolog.Event) {
	switch f.kind {
	case fieldStr:
		e.Str(f.key, f.strVal)
	case fieldInt:
		e.Int(f.key, f.intVal)
	case fieldInt64:
		e.Int64(f.key, f.int64Val)
	case fieldFloat64:
		e.Float64(f.key, f.floatVal)
	case fieldBool:
		e.Bool(f.key, f.boolVal)
	case fieldDur:
		e.Dur(f.key, f.durVal)
	case fieldErr:
		if f.errVal != nil {
			e.Err(f.errVal)
		}
	case fieldAny:
		e.Any(f.key, f.anyVal)
	}
}

func Str(key, val string) Field {
	return Field{kind: fieldStr, key: key, strVal: val}
}

func Int(key string, val int) Field {
	return Field{kind: fieldInt, key: key, intVal: val}
}

func Int64(key string, val int64) Field {
	return Field{kind: fieldInt64, key: key, int64Val: val}
}

func Float64(key string, val float64) Field {
	return Field{kind: fieldFloat64, key: key, floatVal: val}
}

func Bool(key string, val bool) Field {
	return Field{kind: fieldBool, key: key, boolVal: val}
}

func Dur(key string, val time.Duration) Field {
	return Field{kind: fieldDur, key: key, durVal: val}
}

func Err(err error) Field {
	return Field{kind: fieldErr, key: "error", errVal: err}
}

func Any(key string, val any) Field {
	return Field{kind: fieldAny, key: key, anyVal: val}
}
