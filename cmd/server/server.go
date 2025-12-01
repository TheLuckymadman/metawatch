package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/TheLuckymadman/metawatch/internal/config/serverconfig"
	"github.com/TheLuckymadman/metawatch/internal/handler"
	"github.com/TheLuckymadman/metawatch/internal/repository"
	"github.com/TheLuckymadman/metawatch/internal/service"
)

func run() error {
	cfg := serverconfig.Load()
	var sugar zap.SugaredLogger
	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer func() {
		_ = logger.Sync()
	}()
	sugar = *logger.Sugar()

	var s service.Storage
	mode := "FileStorage"
	if cfg.DatabseDSN != "" {
		mode = "Database"
		s, err = repository.NewPGDB(cfg.DatabseDSN, cfg.DBInitMode)
		if err != nil {
			//log.Fatalf("DB connection failed: %v", err)
			return fmt.Errorf("connect database:%w", err)
		}
		defer func() {
			if err := s.Close(); err != nil {
				sugar.Errorf("close storage: %v", err)
			}
		}()
	} else {
		s, err = repository.NewFileStorage(cfg.FileStoragePath, cfg.StoreInterval, cfg.Restore)
		if err != nil {
			return fmt.Errorf("create file storage:%w", err)
		}
		defer func() {
			if err := s.Close(); err != nil {
				sugar.Errorf("close storage: %v", err)
			}
		}()
	}

	srv := service.NewService(s)
	r := chi.NewRouter()
	//r.Use(middleware.RedirectSlashes)
	r.Post("/update/{type}/*", handler.MiddlewareConveyor(handler.MetricSetterHandler(srv), handler.LoggerWrapper(sugar), handler.HashWrapper(cfg.Key)))
	r.Post("/update/", handler.MiddlewareConveyor(handler.JSONSetterHandler(srv), handler.LoggerWrapper(sugar), handler.CompressWrapper, handler.HashWrapper(cfg.Key)))
	r.Post("/updates/", handler.MiddlewareConveyor(handler.JSONSetterHandler(srv), handler.LoggerWrapper(sugar), handler.CompressWrapper, handler.HashWrapper(cfg.Key)))
	r.Get("/value/*", handler.MiddlewareConveyor(handler.MetricGetterHandler(srv), handler.LoggerWrapper(sugar), handler.HashWrapper(cfg.Key), handler.HashWrapper(cfg.Key)))
	r.Post("/value/", handler.MiddlewareConveyor(handler.JSONGetterHandler(srv), handler.LoggerWrapper(sugar), handler.CompressWrapper, handler.HashWrapper(cfg.Key)))
	r.Get("/", handler.MiddlewareConveyor(handler.MetricsListHandler(srv), handler.LoggerWrapper(sugar), handler.CompressWrapper, handler.HashWrapper(cfg.Key)))
	r.Get("/ping", handler.MiddlewareConveyor(handler.PingDB(srv), handler.LoggerWrapper(sugar), handler.HashWrapper(cfg.Key)))

	//log.Printf("Start server on %v", a)

	server := &http.Server{
		Addr:              cfg.ServerURL,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	sugar.Infow(
		"starting server",
		"addr",
		cfg.ServerURL,
		"mode",
		mode,
		"db init mode",
		cfg.DBInitMode,
	)

	errCh := make(chan error, 1)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-stopCtx.Done():
		{
			sugar.Infow("shutting down...")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := server.Shutdown(shutdownCtx); err != nil {
				_ = server.Close()
				return fmt.Errorf("server shutdown: %w", err)
			}
			sugar.Infow("server stopped cleanly")
			return nil
		}
	case err := <-errCh:
		return fmt.Errorf("server:%w", err)
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
