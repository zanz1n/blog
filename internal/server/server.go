package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/zanz1n/blog/internal/repository"
	"github.com/zanz1n/blog/internal/utils/errutils"
	"github.com/zanz1n/blog/internal/utils/xhttp"
	"github.com/zanz1n/blog/web/templates"
)

type Server struct {
	users *repository.UserRepository
	auth  *repository.AuthRepository
}

func New(
	users *repository.UserRepository,
	auth *repository.AuthRepository,
) *Server {
	return &Server{
		users: users,
		auth:  auth,
	}
}

func (s *Server) Wire(r chi.Router) {
	s.wireAuth(r)
}

func (s *Server) Error(w http.ResponseWriter, r *http.Request, err error) {
	token, _ := s.getAuth(w, r)

	xhttp.Error(w, r, templates.PageData[error]{
		Name:  "Blog",
		Token: token,
		Data:  err,
	})
}

func (s *Server) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	token, _ := s.getAuth(w, r)

	data := templates.PageData[error]{
		Name:  "Blog",
		Token: token,
		Data: errutils.NewHttpS(
			"Not Found",
			http.StatusNotFound,
			http.StatusNotFound,
			true,
		),
	}
	xhttp.Error(w, r, data)
}

// Returned http.Cookie can be nil.
func getCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (s *Server) m(h xhttp.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err != nil {
			s.Error(w, r, err)
		}
	}
}

func (s *Server) cfm(
	h xhttp.HandlerFunc,
	full xhttp.ComponentFunc[templates.PageData[error]],
	partial xhttp.ComponentFunc[error],
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err != nil {
			herr := errutils.Http(err)

			if herr.Transparent() {
				errs := strings.ReplaceAll(herr.Error(), "\n", "<br/>")
				err = errors.New(errs)
			} else {
				err = errors.New(http.StatusText(herr.HttpStatus()))
			}

			if xhttp.IsHtmx(r) {
				xhttp.Component(w, r, partial, err, http.StatusOK)
			} else {
				token, _ := s.getAuth(w, r)
				data := templates.PageData[error]{
					Name:  "Blog",
					Token: token,
					Data:  err,
				}

				xhttp.Component(w, r, full, data, herr.HttpStatus())
			}
		}
	}
}
