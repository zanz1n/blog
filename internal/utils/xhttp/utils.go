package xhttp

import (
	"net/http"

	"github.com/zanz1n/blog/internal/utils/errutils"
	"github.com/zanz1n/blog/web/templates"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func ErrorMiddleware(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			Error(w, r, err)
		}
	}
}

func Error(w http.ResponseWriter, r *http.Request, err error) {
	errd := errutils.Http(err)
	data := templates.ErrorData{
		Code:       errd.ErrorCode(),
		HttpStatus: errd.HttpStatus(),
	}

	if errd.Transparent() {
		data.Message = http.StatusText(data.HttpStatus)
	} else {
		data.Message = errd.Error()
	}

	handler(w, r, templates.ErrorPage, data, data.HttpStatus, true)
}

func Redirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	w.Header().Add("HX-Redirect", url)
	http.Redirect(w, r, url, code)
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	Error(w, r,
		errutils.NewHttpS(
			"Page not found",
			http.StatusNotFound,
			http.StatusNotFound,
			true,
		))
}
