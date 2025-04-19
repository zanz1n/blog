package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zanz1n/blog/config"
	"github.com/zanz1n/blog/internal/repository"
	"github.com/zanz1n/blog/internal/server"
)

var (
	jsonLogs = flag.Bool(
		"json-logs",
		false,
		"enables or disables json logging",
	)
	debugLogs = flag.Bool("debug", false, "enables debug logs")
)

var interrupt = make(chan os.Signal, 1)

func init() {
	flag.Parse()

	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.Get()
	if err != nil {
		fatal(err)
	}

	if *debugLogs {
		cfg.LogLevel = slog.LevelDebug
	}

	if *jsonLogs {
		slog.SetDefault(slog.New(slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: cfg.LogLevel},
		)))
	} else {
		slog.SetLogLoggerLevel(cfg.LogLevel)
	}
}

func main() {
	fmt.Println("Running", config.Name, config.Version)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-interrupt
		cancel()
	}()

	if err := main2(ctx); err != nil {
		fatal(err)
	}
}

func main2(ctx context.Context) error {
	cfg, err := config.Get()
	if err != nil {
		return err
	}

	db, err := dbconnect(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	kv, err := kvconnect(db)
	if err != nil {
		return err
	}
	defer kv.Close()

	userRepo := repository.NewUserRepository(db)
	defer userRepo.Close()

	articlesRepo := repository.NewArticleRepository(db)
	defer articlesRepo.Close()

	jwtPub, jwtPriv, err := jwtKeyPair()
	if err != nil {
		return err
	}

	authRepo := repository.NewAuthRepository(jwtPriv, jwtPub, "SRV", kv)

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

	s := server.New(userRepo, authRepo, cfg)

	r.NotFound(s.NotFoundHandler())
	s.Wire(r)

	return listen(ctx, r)
}
