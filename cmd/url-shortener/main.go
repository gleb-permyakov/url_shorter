package main

import (
	// "fmt"
	"log/slog"
	"main/internal/config"
	"main/internal/http-server/handlers/redirect"
	"main/internal/http-server/handlers/url/save"
	"main/internal/http-server/middleware/logger"
	"main/internal/lib/logger/sl"
	"main/internal/storage"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// TODO: init config: cleanenv
	cfg := config.MustLoad()

	// TODO: init logger: slog - смая актуальная библиотека для работы с логами
	log := setupLogger(cfg.Env)

	log.Info("starting url-shortener", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	// TODO: init storage: postgresql
	storage, err := storage.New()
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	defer storage.CloseDb()

	// TODO: init router: chi (совместим с пакетом net/http), "chi render"
	router := chi.NewRouter()

	// middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat) 

	router.Post("/url", save.New(log, storage))
	router.Get("/{alias}", redirect.New(log, storage)) // моем обратиться к alias благодаря middleware.URLFormat

	log.Info("starting server", slog.String("address", cfg.Serv.Address))

	srv := &http.Server{
		Addr: cfg.Serv.Address,
		Handler: router,
		ReadTimeout: cfg.Serv.Timeout,
		WriteTimeout: cfg.Serv.Timeout,
		IdleTimeout: cfg.Serv.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	// TODO: run server
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		TextHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
		log = slog.New(TextHandler)
	case envDev:
		JSONHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
		log = slog.New(JSONHandler)
	case envProd:
		JSONHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
		log = slog.New(JSONHandler)
	}

	return log
}
