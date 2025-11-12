package main

import (
	"log"
	"net/http"

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
		panic(err)
	}
	defer logger.Sync()

	sugar = *logger.Sugar()


	s := repository.NewFileStorage(cfg.FileStoragePath, cfg.StoreInterval, cfg.Restore)
	srv := service.NewService(s)
	r := chi.NewRouter()
	//r.Use(middleware.RedirectSlashes)
	r.Post("/update/{type}/*", handler.MiddlewareConveyor(handler.MetricSetterHandler(srv), handler.LoggerWrapper(sugar)))
	r.Post("/update/", handler.MiddlewareConveyor(handler.JSONSetterHandler(srv), handler.LoggerWrapper(sugar), handler.CompressWrapper))
	r.Get("/value/*", handler.MiddlewareConveyor(handler.MetricGetterHandler(srv), handler.LoggerWrapper(sugar)))
	r.Post("/value/", handler.MiddlewareConveyor(handler.JSONGetterHandler(srv), handler.LoggerWrapper(sugar), handler.CompressWrapper))
	r.Get("/", handler.MiddlewareConveyor(handler.MetricsListHandler(srv), handler.LoggerWrapper(sugar), handler.CompressWrapper))
	
	if cfg.DatabseDSN  != "" {
		db, err := repository.NewPGDB(cfg.DatabseDSN)
		if err != nil {
			log.Fatalf("DB connection filed: %v", err)
		}
		defer db.Close()
		srvDB := service.NewService(db)
		r.Get("/ping", handler.MiddlewareConveyor(handler.PingDB(srvDB), handler.LoggerWrapper(sugar)))
	}
	
	//log.Printf("Start server on %v", a)
	sugar.Infow(
		"Starting server",
		"addr",
		cfg.ServerURL,
	)

	return http.ListenAndServe(cfg.ServerURL, r)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
