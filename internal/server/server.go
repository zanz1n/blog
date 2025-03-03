package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zanz1n/blog/internal/repository"
	"github.com/zanz1n/blog/internal/utils/xhttp"
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

func m(h xhttp.HandlerFunc) http.HandlerFunc {
	return xhttp.ErrorMiddleware(h)
}
