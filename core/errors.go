package core

import "fmt"

type Error struct {
	originalError error
	extraInfo     string
}

func NewError(format string, a ...interface{}) *Error {
	return &Error{
		originalError: fmt.Errorf(format, a...),
		extraInfo:     "",
	}
}

func NewErrorExtraInfo(originalError error, extraInfo string) *Error {
	return &Error{
		originalError: originalError,
		extraInfo:     extraInfo,
	}
}

func (e *Error) Error() string {
	return e.originalError.Error()
}

func (e *Error) ExtraInfo() string {
	return e.extraInfo
}

func (e *Error) OriginalError() error {
	return e.originalError
}
