package base

import "fmt"

type Error struct {
	Code       int
	Message    string
	HTTPStatus int
}

func (err *Error) Error() string { // 指针接收者
	return err.Message
}

func NewError(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

func NewErrorf(code int, format string, a ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, a...)}
}

func NewErrorWithStatus(code int, msg string, status int) *Error {
	return &Error{
		Code:       code,
		Message:    msg,
		HTTPStatus: status,
	}
}
