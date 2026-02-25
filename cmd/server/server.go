package main

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "net/http/pprof"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/TheLuckymadman/metawatch/internal/config/serverconfig"
	"github.com/TheLuckymadman/metawatch/internal/crypto"
	"github.com/TheLuckymadman/metawatch/internal/handler"
	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/TheLuckymadman/metawatch/internal/repository"
	"github.com/TheLuckymadman/metawatch/internal/service"
)

var buildVersion string
var buildDate string
var buildCommit string

func run(stopCtx context.Context) error {
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

	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	sugar.Info("Build version: ", buildVersion)
	sugar.Info("Build date: ", buildDate)
	sugar.Info("Build commit: ", buildCommit)

	var s service.Storage
	mode := "FileStorage"
	if cfg.DatabaseDSN != "" {
		mode = "Database"
		s, err = repository.NewPGDB(cfg.DatabaseDSN, cfg.DBInitMode)
		if err != nil {
			//log.Fatalf("DB connection failed: %v", err)
			return fmt.Errorf("connect database:%w", err)
		}
		defer func() {
			if serr := s.Close(); serr != nil {
				sugar.Errorf("close storage: %v", serr)
			}
		}()
	} else {
		s, err = repository.NewFileStorage(cfg.FileStoragePath, time.Duration(cfg.StoreInterval), cfg.Restore)
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
	if cfg.AuditFile != "" {
		a, err := service.NewAudit(model.AuditToFile, model.AuditMsg{}, cfg.AuditFile, "")
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		srv.Register(a)
		sugar.Infow("Audit enabled", "type", "file", "file path", cfg.AuditFile)
	}
	if cfg.AuditURL != "" {
		a, err := service.NewAudit(model.AuditToServer, model.AuditMsg{}, "", cfg.AuditURL)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		srv.Register(a)
		sugar.Infow("Audit enabled", "type", "server", "url", cfg.AuditURL)
	}

	var privKey *rsa.PrivateKey
	if cfg.CryptoKey != "" {
		privKey, err = crypto.ReadPrivKey(cfg.CryptoKey)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	r := chi.NewRouter()
	r.Post("/update/{type}/*", handler.MiddlewareConveyor(handler.MetricSetterHandler(srv), handler.LoggerWrapper(sugar), handler.HashWrapper(cfg.Key)))
	r.Post("/update/", handler.MiddlewareConveyor(handler.JSONSetterHandler(srv), handler.LoggerWrapper(sugar), handler.DecryptWrapper(privKey), handler.CompressWrapper, handler.HashWrapper(cfg.Key)))
	r.Post("/updates/", handler.MiddlewareConveyor(handler.JSONSetterHandler(srv), handler.LoggerWrapper(sugar), handler.DecryptWrapper(privKey), handler.CompressWrapper, handler.HashWrapper(cfg.Key)))
	r.Get("/value/*", handler.MiddlewareConveyor(handler.MetricGetterHandler(srv), handler.LoggerWrapper(sugar), handler.CompressWrapper, handler.HashWrapper(cfg.Key)))
	r.Post("/value/", handler.MiddlewareConveyor(handler.JSONGetterHandler(srv), handler.LoggerWrapper(sugar), handler.DecryptWrapper(privKey), handler.CompressWrapper, handler.HashWrapper(cfg.Key)))
	r.Get("/", handler.MiddlewareConveyor(handler.MetricsListHandler(srv), handler.LoggerWrapper(sugar), handler.CompressWrapper, handler.HashWrapper(cfg.Key)))
	r.Get("/ping", handler.MiddlewareConveyor(handler.PingDB(srv), handler.LoggerWrapper(sugar), handler.HashWrapper(cfg.Key)))

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
		"StoreInterval",
		time.Duration(cfg.StoreInterval).Seconds(),
		"FileStoragePath",
		cfg.FileStoragePath,
		"Restore",
		cfg.Restore,
	)

	errCh := make(chan error, 1)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

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
	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	g, ctx := errgroup.WithContext(stopCtx)

	pprofsvr := http.Server{
		Addr: "localhost:6060",
	}

	g.Go(func() error {
		log.Println("pprof listening on localhost:6060")
		return pprofsvr.ListenAndServe()
	})

	g.Go(func() error {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		err := pprofsvr.Shutdown(shutCtx)
		if err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		if err := run(ctx); err != nil {
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
