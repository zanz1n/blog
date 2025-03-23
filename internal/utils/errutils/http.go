package errutils

import (
	"errors"
	"net/http"
)

type HttpError interface {
	HttpStatus() int
	ErrorCode() int32
	Transparent() bool
	error
}

type httpStatus struct {
	code        int32
	status      int16
	transparent bool
	error
}

func (e *httpStatus) HttpStatus() int {
	return int(e.status)
}

func (e *httpStatus) ErrorCode() int32 {
	return e.code
}

func (e *httpStatus) Transparent() bool {
	return e.transparent
}

func NewHttpS(err string, status int16, code int32, transparent ...bool) HttpError {
	return NewHttp(errors.New(err), status, code, transparent...)
}

func NewHttp(err error, status int16, code int32, transparent ...bool) HttpError {
	if err == nil {
		err = errors.New("unknown error")
		transparent = []bool{true}
	}

	e := &httpStatus{
		code:   code,
		status: status,
		error:  err,
	}

	if len(transparent) >= 1 {
		e.transparent = transparent[0]
	}

	return e
}

func Http(err error) HttpError {
	if err == nil {
		return nil
	} else if err, ok := err.(HttpError); ok {
		return err
	}

	return &httpStatus{
		code:        0,
		status:      http.StatusInternalServerError,
		transparent: false,
		error:       err,
	}
}
