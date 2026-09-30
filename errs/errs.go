package errs

import (
	"errors"
	"fmt"
)

const defaultStatus = 200

type Error struct {
	code   int
	msg    string
	status int
	cause  error
}

func New(code int, msg string) *Error {
	return NewWithStatus(code, msg, defaultStatus)
}

func NewWithStatus(code int, msg string, status int) *Error {
	if status == 0 {
		status = defaultStatus
	}
	return &Error{code: code, msg: msg, status: status}
}

var ErrUnclassified = NewWithStatus(-1, "internal error", 500)

func (e *Error) Wrap(cause error) *Error {
	if cause == nil {
		return e
	}
	c := *e
	c.cause = cause
	return &c
}

func (e *Error) Wrapf(format string, a ...any) *Error {
	return e.Wrap(fmt.Errorf(format, a...))
}

func (e *Error) With(format string, a ...any) *Error {
	c := *e
	c.msg = fmt.Sprintf(format, a...)
	return &c
}

func (e *Error) Unwrap() error { return e.cause }

func (e *Error) Code() int       { return e.code }
func (e *Error) Message() string { return e.msg }
func (e *Error) Status() int     { return e.status }
func (e *Error) Cause() error    { return e.cause }

func (e *Error) Error() string {
	if e.cause == nil {
		return e.msg
	}
	return e.msg + ": " + e.cause.Error()
}

func (e *Error) Is(target error) bool {
	var t *Error
	ok := errors.As(target, &t)
	return ok && t.code == e.code
}

func As(err error) (*Error, bool) {
	return errors.AsType[*Error](err)
}

func IsCode(err error, code int) bool {
	e, ok := As(err)
	return ok && e.code == code
}

func Normalize(err error) *Error {
	if err == nil {
		return nil
	}
	if e, ok := As(err); ok {
		return e
	}
	return ErrUnclassified.Wrap(err)
}
