package server

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/repository"
	"github.com/zanz1n/blog/internal/utils/errutils"
	"github.com/zanz1n/blog/internal/utils/xhttp"
	"github.com/zanz1n/blog/web/templates"
)

type SignUpRequest = dto.UserCreateData

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=128"`
	Password string `json:"password" validate:"required,max=256"`
}

var (
	ErrUnauthorized = errutils.NewHttpS(
		"User not found or password doesn't match",
		http.StatusUnauthorized,
		http.StatusUnauthorized,
		true,
	)
)

func (s *Server) wireAuth(r chi.Router) {
	r.Get("/auth/signup", s.m(s.GetAuthSignup))
	r.Post("/auth/signup", s.cfm(
		s.PostAuthSignup,
		templates.SignUpPage,
		templates.SignUpForm,
	))

	r.Get("/auth/login", s.m(s.GetAuthLogin))
	r.Post("/auth/login", s.cfm(
		s.PostAuthLogin,
		templates.LoginPage,
		templates.LoginForm,
	))

	r.Get("/auth/logout", s.m(s.GetAuthLogout))
}

func (s *Server) GetAuthSignup(c *xhttp.Ctx) error {
	token, _ := c.GetAuth()
	data := templates.PageData[error]{
		Name:  "Blog",
		Token: token,
		Data:  nil,
	}

	return xhttp.Component(c, templates.SignUpPage, data, http.StatusOK)
}

func (s *Server) PostAuthSignup(c *xhttp.Ctx) error {
	var data SignUpRequest
	err := c.Parse(&data)
	if err != nil {
		return err
	}

	user, err := dto.NewUser(data, dto.PermissionDefault, s.cfg.BcryptCost)
	if err != nil {
		return err
	}

	if err = s.users.Create(c.Context(), user); err != nil {
		return err
	}

	refreshToken, err := s.auth.GenRefreshToken(c.Context(), user.ID)
	if err != nil {
		return err
	}

	token := dto.NewAuthToken(&user, "", s.cfg.JWT.GetDuration())
	authToken, err := s.auth.EncodeToken(token)
	if err != nil {
		return err
	}

	c.SetCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: refreshToken,
		Path:  "/",
	})
	c.SetCookie(&http.Cookie{
		Name:    "auth_token",
		Value:   authToken,
		Path:    "/",
		Expires: token.ExpiresAt.Time,
	})

	c.Redirect("/")
	return nil
}

func (s *Server) GetAuthLogin(c *xhttp.Ctx) error {
	token, _ := c.GetAuth()
	data := templates.PageData[error]{
		Name:  "Blog",
		Token: token,
		Data:  nil,
	}

	return xhttp.Component(c, templates.LoginPage, data, http.StatusOK)
}

func (s *Server) PostAuthLogin(c *xhttp.Ctx) error {
	var data LoginRequest
	err := c.Parse(&data)
	if err != nil {
		return err
	}

	user, err := s.users.GetByEmail(c.Context(), data.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			err = ErrUnauthorized
		}
		return err
	}

	if !user.PasswordMatches(data.Password) {
		return ErrUnauthorized
	}

	refreshToken, err := s.auth.GenRefreshToken(c.Context(), user.ID)
	if err != nil {
		return err
	}

	token := dto.NewAuthToken(&user, "", s.cfg.JWT.GetDuration())
	authToken, err := s.auth.EncodeToken(token)
	if err != nil {
		return err
	}

	c.SetCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: refreshToken,
		Path:  "/",
	})
	c.SetCookie(&http.Cookie{
		Name:    "auth_token",
		Value:   authToken,
		Expires: token.ExpiresAt.Time,
		Path:    "/",
	})

	c.Redirect("/")
	return nil
}

func (s *Server) GetAuthLogout(c *xhttp.Ctx) error {
	c.DelCookie("refresh_token")
	c.DelCookie("auth_token")

	c.Redirect("/")
	return nil
}
