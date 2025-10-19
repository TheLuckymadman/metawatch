package main

import (
	"flag"
	"net/http"
	"log"
	
	"github.com/go-chi/chi/v5"
	//"github.com/go-chi/chi/v5/middleware"

	"github.com/TheLuckymadman/metawatch/internal/handler"
	"github.com/TheLuckymadman/metawatch/internal/repository"
)

var (
	a = flag.String("a", "localhost:8080", "local listening interface in the format servername:port")
)

func run() error {
	s := repository.NewStorage()
	r := chi.NewRouter()
	//r.Use(middleware.RedirectSlashes)
	r.Post("/update/{type}/*", handler.MetricReceiverHandler(s))
	r.Get("/value/*", handler.MetricGetterHandler(s))
	r.Get("/", handler.MetricsGetterHandler(s))

	log.Printf("Start server on %v", *a)

	return http.ListenAndServe(*a, r)
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		panic(err)
	}
}
