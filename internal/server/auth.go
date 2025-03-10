package server

import (
	"errors"
	"net/http"
	"time"

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

	r.Get("/auth/logout", s.GetAuthLogout)
}

func (s *Server) getAuth(w http.ResponseWriter, r *http.Request) (*dto.AuthToken, error) {
	cookies := r.Cookies()
	authToken := getCookie(cookies, "auth_token")
	if authToken == nil {
		refreshToken := getCookie(cookies, "refresh_token")
		if refreshToken == nil {
			return nil, nil
		}

		userId, err := s.auth.ValidateRefreshToken(
			r.Context(),
			refreshToken.Value,
		)
		if err != nil {
			return nil, err
		}

		user, err := s.users.GetById(r.Context(), userId)
		if err != nil {
			return nil, err
		}

		// TODO: add jwt duration to configuration
		token := dto.NewAuthToken(&user, "", time.Hour)
		authToken, err := s.auth.EncodeToken(token)
		if err != nil {
			return nil, err
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "auth_token",
			Value:   authToken,
			Path:    "/",
			Expires: token.ExpiresAt.Time,
		})
		return &token, nil
	}

	token, err := s.auth.DecodeToken(authToken.Value)
	return &token, err
}

func (s *Server) GetAuthSignup(w http.ResponseWriter, r *http.Request) error {
	token, _ := s.getAuth(w, r)
	data := templates.PageData[error]{
		Name:  "Blog",
		Token: token,
		Data:  nil,
	}

	return xhttp.Component(w, r, templates.SignUpPage, data, http.StatusOK)
}

func (s *Server) PostAuthSignup(w http.ResponseWriter, r *http.Request) error {
	data, err := xhttp.Parse[SignUpRequest](r)
	if err != nil {
		return err
	}

	// TODO: add hashing cost to configuration
	user, err := dto.NewUser(data, dto.PermissionDefault, 12)
	if err != nil {
		return err
	}

	if err = s.users.Create(r.Context(), user); err != nil {
		return err
	}

	refreshToken, err := s.auth.GenRefreshToken(r.Context(), user.ID)
	if err != nil {
		return err
	}

	// TODO: add jwt duration to configuration
	token := dto.NewAuthToken(&user, "", time.Hour)
	authToken, err := s.auth.EncodeToken(token)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "refresh_token",
		Value: refreshToken,
		Path:  "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "auth_token",
		Value:   authToken,
		Path:    "/",
		Expires: token.ExpiresAt.Time,
	})

	xhttp.Redirect(w, r, "/")
	return nil
}

func (s *Server) GetAuthLogin(w http.ResponseWriter, r *http.Request) error {
	token, _ := s.getAuth(w, r)
	data := templates.PageData[error]{
		Name:  "Blog",
		Token: token,
		Data:  nil,
	}

	return xhttp.Component(w, r, templates.LoginPage, data, http.StatusOK)
}

func (s *Server) PostAuthLogin(w http.ResponseWriter, r *http.Request) error {
	data, err := xhttp.Parse[LoginRequest](r)
	if err != nil {
		return err
	}

	user, err := s.users.GetByEmail(r.Context(), data.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			err = ErrUnauthorized
		}
		return err
	}

	if !user.PasswordMatches(data.Password) {
		return ErrUnauthorized
	}

	refreshToken, err := s.auth.GenRefreshToken(r.Context(), user.ID)
	if err != nil {
		return err
	}

	// TODO: add jwt duration to configuration
	token := dto.NewAuthToken(&user, "", time.Hour)
	authToken, err := s.auth.EncodeToken(token)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "refresh_token",
		Value: refreshToken,
		Path:  "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "auth_token",
		Value:   authToken,
		Expires: token.ExpiresAt.Time,
		Path:    "/",
	})

	xhttp.Redirect(w, r, "/")
	return nil
}

func (s *Server) GetAuthLogout(w http.ResponseWriter, r *http.Request) {
	xhttp.DelCookie(w, "refresh_token")
	xhttp.DelCookie(w, "auth_token")

	xhttp.Redirect(w, r, "/")
}
