package main

import (
	"net/http"
	//"log"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/TheLuckymadman/metawatch/internal/config/serverconfig"
	"github.com/TheLuckymadman/metawatch/internal/handler"
	"github.com/TheLuckymadman/metawatch/internal/repository"
)

var (
	a string
)

func run() error {
	cfg := serverconfig.Load()
	a = (*cfg).ServerURL

	var sugar zap.SugaredLogger
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	sugar = *logger.Sugar()

	s := repository.NewStorage()
	r := chi.NewRouter()
	//r.Use(middleware.RedirectSlashes)
	r.Post("/update/{type}/*", handler.MiddlewareConveyor(handler.MetricSetterHandler(s), handler.LoggerWrapper(sugar)))
	r.Post("/update/", handler.MiddlewareConveyor(handler.JSONSetterHandler(s), handler.LoggerWrapper(sugar)))
	r.Get("/value/*", handler.MiddlewareConveyor(handler.MetricGetterHandler(s), handler.LoggerWrapper(sugar)))
	r.Post("/value/", handler.MiddlewareConveyor(handler.JSONGetterHandler(s), handler.LoggerWrapper(sugar)))
	r.Get("/", handler.MiddlewareConveyor(handler.MetricsListHandler(s), handler.LoggerWrapper(sugar)))

	//log.Printf("Start server on %v", a)
	sugar.Infow(
		"Starting server",
		"addr",
		a,
	)

	return http.ListenAndServe(a, r)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
