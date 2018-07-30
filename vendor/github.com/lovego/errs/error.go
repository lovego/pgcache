package errs

import (
	"fmt"
)

type Error struct {
	err           error // original error
	stack         string
	code, message string
	data          interface{}
}

func Trace(err error) *Error {
	if erro, ok := err.(*Error); ok {
		return erro
	} else {
		return &Error{err: err, stack: Stack(3)}
	}
}

func Tracef(format string, args ...interface{}) *Error {
	return &Error{err: fmt.Errorf(format, args...), stack: Stack(3)}
}

func New(code, message string) *Error {
	return &Error{code: code, message: message}
}

func (err *Error) Error() string {
	if err.err != nil {
		return err.err.Error()
	} else {
		return err.code + `: ` + err.message
	}
}

func (err *Error) Err() error {
	return err.err
}

func (err *Error) Stack() string {
	return err.stack
}

func (err *Error) Code() string {
	return err.code
}

func (err *Error) Message() string {
	return err.message
}

func (err *Error) Data() interface{} {
	return err.data
}

func (err *Error) SetErr(erro error) *Error {
	err.err = erro
	return err
}

func (err *Error) Trace() *Error {
	err.stack = Stack(3)
	return err
}

func (err *Error) SetCodeMessage(code, message string) *Error {
	err.code, err.message = code, message
	return err
}

func (err *Error) SetData(data interface{}) *Error {
	err.data = data
	return err
}
