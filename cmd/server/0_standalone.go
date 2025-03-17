//go:build !lambda
// +build !lambda

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zanz1n/blog/internal/utils"
)

func init() {
	godotenv.Load()
}

func listen(ctx context.Context, h http.Handler) error {
	server := &http.Server{
		Addr:    os.Getenv("LISTEN_ADDR"),
		Handler: h,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	var shutdownStart time.Time

	go func() {
		<-ctx.Done()
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
