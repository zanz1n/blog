package xhttp

import (
	"net/http"

	"github.com/zanz1n/blog/internal/utils/errutils"
	"github.com/zanz1n/blog/web/templates"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func Error(w http.ResponseWriter, r *http.Request, p templates.PageData[error]) {
	errd := errutils.Http(p.Data)
	data := templates.ErrorData{
		Code:       errd.ErrorCode(),
		HttpStatus: errd.HttpStatus(),
	}

	if errd.Transparent() {
		data.Message = http.StatusText(data.HttpStatus)
	} else {
		data.Message = errd.Error()
	}

	p2 := templates.PageData[templates.ErrorData]{
		Name:  p.Name,
		Token: p.Token,
		Data:  data,
	}

	handler(w, r, templates.ErrorPage, p2, data.HttpStatus, true)
}

func Redirect(w http.ResponseWriter, r *http.Request, url string) {
	if IsHtmx(r) {
		w.Header().Add("HX-Redirect", url)
		w.WriteHeader(http.StatusOK)
	} else {
		http.Redirect(w, r, url, http.StatusFound)
	}
}

func DelCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:   name,
		Path:   "/",
		MaxAge: -1,
	})
}
