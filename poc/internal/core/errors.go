package core

import "fmt"

type Code string

const (
	CodeOK                 Code = "OK"
	CodeInvalidFrame       Code = "INVALID_FRAME"
	CodeUnsupportedVersion Code = "UNSUPPORTED_VERSION"
	CodeUnknownProfile     Code = "UNKNOWN_PROFILE"
	CodeInvalidEnvelope    Code = "INVALID_ENVELOPE"
	CodeInternalError      Code = "INTERNAL_ERROR"
)

type Error struct {
	Code Code
	Err  error
}

func (e *Error) Error() string {
	if e.Err == nil {
		return string(e.Code)
	}
	return fmt.Sprintf("%s: %v", e.Code, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func Wrap(code Code, err error) error {
	if err == nil {
		return &Error{Code: code}
	}
	return &Error{Code: code, Err: err}
}

func CodeFromError(err error) Code {
	if err == nil {
		return CodeOK
	}
	e, ok := err.(*Error)
	if !ok {
		return CodeInternalError
	}
	return e.Code
}
