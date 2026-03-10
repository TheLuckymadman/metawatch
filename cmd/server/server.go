package main

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "net/http/pprof"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/TheLuckymadman/metawatch/internal/config/serverconfig"
	"github.com/TheLuckymadman/metawatch/internal/crypto"
	gr "github.com/TheLuckymadman/metawatch/internal/grpc"
	"github.com/TheLuckymadman/metawatch/internal/handler"
	"github.com/TheLuckymadman/metawatch/internal/handler/middleware"
	"github.com/TheLuckymadman/metawatch/internal/model"
	pb "github.com/TheLuckymadman/metawatch/internal/proto"
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

	var subnet *net.IPNet
	if cfg.TrustedSubnet != "" {
		_, subnet, err = net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			return fmt.Errorf("trusted subnet error: %w", err)
		}
	}

	r := chi.NewRouter()
	r.Post("/update/{type}/*", middleware.MiddlewareConveyor(handler.MetricSetterHandler(srv), middleware.LoggerWrapper(sugar), middleware.SubnetChecker(subnet), middleware.HashWrapper(cfg.Key)))
	r.Post("/update/", middleware.MiddlewareConveyor(handler.JSONSetterHandler(srv), middleware.LoggerWrapper(sugar), middleware.SubnetChecker(subnet), middleware.DecryptWrapper(privKey), middleware.CompressWrapper, middleware.HashWrapper(cfg.Key)))
	r.Post("/updates/", middleware.MiddlewareConveyor(handler.JSONSetterHandler(srv), middleware.LoggerWrapper(sugar), middleware.SubnetChecker(subnet), middleware.DecryptWrapper(privKey), middleware.CompressWrapper, middleware.HashWrapper(cfg.Key)))
	r.Get("/value/*", middleware.MiddlewareConveyor(handler.MetricGetterHandler(srv), middleware.LoggerWrapper(sugar), middleware.CompressWrapper, middleware.HashWrapper(cfg.Key)))
	r.Post("/value/", middleware.MiddlewareConveyor(handler.JSONGetterHandler(srv), middleware.LoggerWrapper(sugar), middleware.DecryptWrapper(privKey), middleware.CompressWrapper, middleware.HashWrapper(cfg.Key)))
	r.Get("/", middleware.MiddlewareConveyor(handler.MetricsListHandler(srv), middleware.LoggerWrapper(sugar), middleware.CompressWrapper, middleware.HashWrapper(cfg.Key)))
	r.Get("/ping", middleware.MiddlewareConveyor(handler.PingDB(srv), middleware.LoggerWrapper(sugar), middleware.HashWrapper(cfg.Key)))

	server := &http.Server{
		Addr:              cfg.ServerURL,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	zapLogFields := []interface{}{
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
		"TrustedSubnet",
		cfg.TrustedSubnet,
	}

	var grpcSrv *grpc.Server
	if cfg.GRPCSrvAddr != "" {
		listen, err := net.Listen("tcp", cfg.GRPCSrvAddr)
		if err != nil {
			return fmt.Errorf("error during GRPC server init: %w", err)
		}
		if subnet != nil {
			grpcSrv = grpc.NewServer(grpc.UnaryInterceptor(gr.SubnetChecker(subnet)))
		} else {
			grpcSrv = grpc.NewServer()
		}
		metricserver := gr.NewMetricServer(srv)
		pb.RegisterMetricsServer(grpcSrv, metricserver)

		go func() {
			if err := grpcSrv.Serve(listen); err != nil {
				errCh <- err
			}
		}()

		zapLogFields = append(zapLogFields, "GRPC is enabled", cfg.GRPCSrvAddr)
	}

	sugar.Infow("starting server", zapLogFields...)

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
			if cfg.GRPCSrvAddr != "" {
				grpcSrv.GracefulStop()
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
