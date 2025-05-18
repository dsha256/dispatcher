package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dsha256/dispatcher/internal/config"
	"github.com/dsha256/dispatcher/internal/dispatcher"
	"github.com/dsha256/dispatcher/internal/handler"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	cfg, err := config.GetConfigFromFile("./config.yaml")
	if err != nil {
		logger.Error("Failed to load config file", "error", err)
		os.Exit(1)
	}

	logger.Info("Starting packer service")

	newDispatcher := dispatcher.New()

	newHandler := handler.New(logger, newDispatcher)

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("Received request",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
			)
			mux := http.NewServeMux()
			newHandler.RegisterRoutes(mux)
			mux.ServeHTTP(w, r)
		}),
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
	}

	go func() {
		logger.Info("Server starting", "port", cfg.Server.Port)
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited properly")
}
