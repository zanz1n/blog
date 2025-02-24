//go:build !lambda
// +build !lambda

package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zanz1n/blog/config"
	"github.com/zanz1n/blog/internal/repository"
	"github.com/zanz1n/blog/internal/server"
	"github.com/zanz1n/blog/internal/utils"
)

var interrupt = make(chan os.Signal, 1)

func init() {
	godotenv.Load()
	flag.Parse()

	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", "file:sqlite.db")
	}

	if os.Getenv("LISTEN_ADDR") == "" {
		os.Setenv("LISTEN_ADDR", ":8080")
	}
}

func main() {
	fmt.Println("Running", config.Name, config.Version)

	if err := main2(); err != nil {
		fatal(err)
	}
}

func main2() error {
	db, err := dbconnect()
	if err != nil {
		return err
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	defer userRepo.Close()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	authRepo := repository.NewAuthRepository(priv, pub, "SRV")

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           86400,
	}))

	if err = wireStaticRoutes(r); err != nil {
		return err
	}

	server.New(userRepo, authRepo).Wire(r)

	return listen(r)
}

func listen(h http.Handler) error {
	server := &http.Server{
		Addr:    os.Getenv("LISTEN_ADDR"),
		Handler: h,
	}

	var shutdownStart time.Time

	go func() {
		<-interrupt
		shutdownStart = time.Now()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			slog.Error("Failed to graceful shutdown server: " + err.Error())
			return
		}
	}()

	slog.Info("HTTP: Listening for connections", "addr", server.Addr)

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to listen http server: %s", err)
	}

	slog.Info(
		"HTTP: Shutted down server",
		utils.TookAttr(shutdownStart, time.Microsecond),
	)

	return nil
}
