package main

import (
	"net/http"

	"github.com/TheLuckymadman/metawatch/internal/repository"
	"github.com/TheLuckymadman/metawatch/internal/handler"
)

func run() error {
	s := repository.NewStorage()
	m := http.NewServeMux()
	m.HandleFunc("/update/", handler.MetricReceiverHandler(s))
	return http.ListenAndServe("localhost:8080", m)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
