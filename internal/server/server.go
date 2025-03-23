package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/zanz1n/blog/config"
	"github.com/zanz1n/blog/internal/repository"
	"github.com/zanz1n/blog/internal/utils/errutils"
	"github.com/zanz1n/blog/internal/utils/xhttp"
	"github.com/zanz1n/blog/web/templates"
)

type Server struct {
	users *repository.UserRepository
	auth  *repository.AuthRepository

	cfg *config.Config
}

func New(
	users *repository.UserRepository,
	auth *repository.AuthRepository,
	cfg *config.Config,
) *Server {
	return &Server{
		users: users,
		auth:  auth,
		cfg:   cfg,
	}
}

func (s *Server) Wire(r chi.Router) {
	s.wireAuth(r)
}

func (s *Server) NotFoundHandler() http.HandlerFunc {
	return s.m(func(c *xhttp.Ctx) error {
		return errutils.NewHttpS(
			"Not Found",
			http.StatusNotFound,
			http.StatusNotFound,
			true,
		)
	})
}

func (s *Server) m(h xhttp.HandlerFunc) http.HandlerFunc {
	return xhttp.CtxHandler(h, s.auth, s.users, s.cfg, true)
}

func (s *Server) cfm(
	h xhttp.HandlerFunc,
	full xhttp.ComponentFunc[templates.PageData[error]],
	partial xhttp.ComponentFunc[error],
) http.HandlerFunc {
	return s.m(func(c *xhttp.Ctx) error {
		err := h(c)
		if err != nil {
			herr := errutils.Http(err)

			if herr.Transparent() {
				errs := strings.ReplaceAll(herr.Error(), "\n", "<br/>")
				err = errors.New(errs)
			} else {
				err = errors.New(http.StatusText(herr.HttpStatus()))
			}

			if c.IsHtmx() {
				xhttp.Component(c, partial, err, http.StatusOK)
			} else {
				token, _ := c.GetAuth()
				data := templates.PageData[error]{
					Name:  "Blog",
					Token: token,
					Data:  err,
				}

				xhttp.Component(c, full, data, herr.HttpStatus())
			}
		}
		return nil
	})
}
