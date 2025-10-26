package main

import (
	"net/http"
	"log"
	
	"github.com/go-chi/chi/v5"
	//"github.com/go-chi/chi/v5/middleware"

	"github.com/TheLuckymadman/metawatch/internal/handler"
	"github.com/TheLuckymadman/metawatch/internal/repository"
	"github.com/TheLuckymadman/metawatch/internal/config/serverconfig"
)

var (
	a string
)

func run() error {
	cfg := serverconfig.Load()
	a = (*cfg).ServerURL

	s := repository.NewStorage()
	r := chi.NewRouter()
	//r.Use(middleware.RedirectSlashes)
	r.Post("/update/{type}/*", handler.MetricReceiverHandler(s))
	r.Get("/value/*", handler.MetricGetterHandler(s))
	r.Get("/", handler.MetricsGetterHandler(s))

	log.Printf("Start server on %v", a)

	return http.ListenAndServe(a, r)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
