package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/TheLuckymadman/metawatch/internal/handler"
	"github.com/TheLuckymadman/metawatch/internal/repository"
)

func run() error {
	s := repository.NewStorage()
	r := chi.NewRouter()
	r.Use(middleware.RedirectSlashes) 
	r.Post(`/update/{type}/*`, handler.MetricReceiverHandler(s))
	r.Get("/value/*", handler.MetricGetterHandler(s))
	r.Get("/", handler.MetricsGetterHandler(s))

	return http.ListenAndServe("localhost:8080", r)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
