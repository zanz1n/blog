package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/utils/errutils"
	"github.com/zanz1n/blog/internal/utils/xhttp"
	"github.com/zanz1n/blog/web/templates"
)

type SignUpRequest = dto.UserCreateData

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=128"`
	Password string `json:"password" validate:"required,min=8,max=256"`
}

func (s *Server) wireAuth(r chi.Router) {
	r.Get("/auth/signup", m(s.GetAuthSignup))
	r.Post("/auth/signup", m(cfm(
		s.PostAuthSignup,
		templates.SignUpPage,
		templates.SignUpForm,
	)))

	r.Get("/auth/login", m(s.GetAuthLogin))
	r.Post("/auth/login", m(cfm(
		s.PostAuthLogin,
		templates.LoginPage,
		templates.LoginForm,
	)))
}

func (s *Server) GetAuthSignup(w http.ResponseWriter, r *http.Request) error {
	return xhttp.Component(w, r, templates.SignUpPage, nil, http.StatusOK)
}

func (s *Server) PostAuthSignup(w http.ResponseWriter, r *http.Request) error {
	_, err := xhttp.Parse[SignUpRequest](r)
	if err != nil {
		return err
	}

	return errutils.NewHttpS(
		"Unimplemented",
		http.StatusNotImplemented,
		http.StatusNotImplemented,
		true,
	)
}

func (s *Server) GetAuthLogin(w http.ResponseWriter, r *http.Request) error {
	return xhttp.Component(w, r, templates.LoginPage, nil, http.StatusOK)
}

func (s *Server) PostAuthLogin(w http.ResponseWriter, r *http.Request) error {
	_, err := xhttp.Parse[LoginRequest](r)
	if err != nil {
		return err
	}

	return errutils.NewHttpS(
		"Unimplemented",
		http.StatusNotImplemented,
		http.StatusNotImplemented,
		true,
	)
}

func cfm(
	h xhttp.HandlerFunc,
	full, partial xhttp.ComponentFunc[error],
) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		err := h(w, r)
		if err != nil {
			herr := errutils.Http(err)

			cf := full
			status := http.StatusOK
			if xhttp.IsHtmx(r) {
				cf = partial
			} else {
				status = herr.HttpStatus()
			}

			if herr.Transparent() {
				errs := strings.ReplaceAll(herr.Error(), "\n", "<br/>")
				return xhttp.Component(w, r, cf, errors.New(errs), status)
			} else {
				return xhttp.Component(w, r, cf,
					errors.New(http.StatusText(herr.HttpStatus())),
					status,
				)
			}
		}
		return nil
	}
}
