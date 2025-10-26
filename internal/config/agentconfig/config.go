package agentconfig

import (
	"log"
	"flag"
	"strings"

	"github.com/caarlos0/env"
)

type Config struct {
	ServerURL string `env:"ADDRESS"`
	PollInterval int `env:"POLL_INTERVAL"`
	ReportInterval int `env:"REPORT_INTERVAL"`
}

func GetDefaultConfig() Config{
	return Config{
	ServerURL: "http://127.0.0.1:8080",
	PollInterval: 2,
	ReportInterval: 10,
	}
}

func Load() Config {
	cfg := GetDefaultConfig()
	flag.StringVar(&cfg.ServerURL, "a", cfg.ServerURL, "server endpoint address in the format http://server:port")
	flag.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "fetching metrics frequency") 
	flag.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "sending metrics frequency")
	
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if !strings.HasPrefix(cfg.ServerURL, "http://") && !strings.HasPrefix(cfg.ServerURL, "https://") {
        cfg.ServerURL = "http://" + cfg.ServerURL
    }

	return  cfg
}
