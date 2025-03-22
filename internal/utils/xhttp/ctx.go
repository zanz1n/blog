package xhttp

import (
	"net/http"
	"time"

	"github.com/zanz1n/blog/internal/dto"
	"github.com/zanz1n/blog/internal/repository"
	"github.com/zanz1n/blog/web/templates"
)

type HandlerFunc func(c *Ctx) error

func CtxHandler(
	h HandlerFunc,
	auth *repository.AuthRepository,
	users *repository.UserRepository,
	logs bool,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		c := newCtx(w, r, auth, users)

		err := h(c)
		if err != nil {
			token, _ := c.GetAuth()
			Error(c, templates.PageData[error]{
				Name:  "Blog",
				Token: token,
				Data:  err,
			})
		}

		if logs {
			LogRequest(start, c, err)
		}
	}
}

func newCtx(
	w http.ResponseWriter,
	r *http.Request,
	auth *repository.AuthRepository,
	users *repository.UserRepository,
) *Ctx {
	return &Ctx{w: w, Request: r, authr: auth, users: users}
}

var _ http.ResponseWriter = &Ctx{}

type Ctx struct {
	w http.ResponseWriter
	*http.Request

	authr *repository.AuthRepository
	users *repository.UserRepository

	statusCode int

	auth       *dto.AuthToken
	authParsed bool

	cookies       []*http.Cookie
	cookiesParsed bool
}

// Header implements http.ResponseWriter.
func (c *Ctx) Header() http.Header {
	return c.w.Header()
}

// Write implements http.ResponseWriter.
func (c *Ctx) Write(b []byte) (int, error) {
	return c.w.Write(b)
}

// WriteHeader implements http.ResponseWriter.
func (c *Ctx) WriteHeader(statusCode int) {
	c.statusCode = statusCode
	c.w.WriteHeader(statusCode)
}

func (c *Ctx) GetStatusCode() int {
	if c.statusCode == 0 {
		return http.StatusOK
	}
	return c.statusCode
}

func (c *Ctx) AddHeader(key, value string) {
	c.Header().Add(key, value)
}

func (c *Ctx) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

func (c *Ctx) Parse(v any) error {
	return parse(c.Request, v)
}

func (c *Ctx) Cookies() []*http.Cookie {
	if !c.cookiesParsed {
		c.cookies = c.Request.Cookies()
	}

	return c.cookies
}

// The cookie can be nil
func (c *Ctx) GetCookie(name string) *http.Cookie {
	cookies := c.Cookies()
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (c *Ctx) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c, cookie)
}

func (c *Ctx) DelCookie(name string) {
	http.SetCookie(c, &http.Cookie{
		Name:   name,
		Path:   "/",
		MaxAge: -1,
	})
}

func (c *Ctx) IsHtmx() bool {
	return c.GetHeader("HX-Request") == "true"
}

func (c *Ctx) Redirect(url string) {
	if c.IsHtmx() {
		c.AddHeader("HX-Redirect", url)
		c.WriteHeader(http.StatusOK)
	} else {
		http.Redirect(c, c.Request, url, http.StatusFound)
	}
}

// The token can be nil
func (c *Ctx) GetAuth() (*dto.AuthToken, error) {
	if c.authParsed {
		return c.auth, nil
	}

	authToken := c.GetCookie("auth_token")
	if authToken == nil {
		refreshToken := c.GetCookie("refresh_token")
		if refreshToken == nil {
			c.authParsed = true
			return nil, nil
		}

		userId, err := c.authr.ValidateRefreshToken(
			c.Context(),
			refreshToken.Value,
		)
		if err != nil {
			return nil, err
		}

		user, err := c.users.GetById(c.Context(), userId)
		if err != nil {
			return nil, err
		}

		// TODO: add jwt duration to configuration
		token := dto.NewAuthToken(&user, "", time.Hour)
		authToken, err := c.authr.EncodeToken(token)
		if err != nil {
			return nil, err
		}

		c.auth = &token
		c.authParsed = true

		c.SetCookie(&http.Cookie{
			Name:    "auth_token",
			Value:   authToken,
			Path:    "/",
			Expires: token.ExpiresAt.Time,
		})
		return &token, nil
	}

	token, err := c.authr.DecodeToken(authToken.Value)
	if err == nil {
		c.auth = &token
		c.authParsed = true
	}
	return &token, err
}
