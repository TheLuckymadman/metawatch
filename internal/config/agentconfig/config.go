package agentconfig

import (
	"flag"
	"log"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerURL      string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	BatchSize      int    `env:"BATCH_SIZE"`
	Compress       bool   `env:"COMPRESS"`
	Key            string `env:"KEY"`
}

func GetDefaultConfig() Config {
	return Config{
		ServerURL:      "http://127.0.0.1:8080",
		PollInterval:   2,
		ReportInterval: 10,
		BatchSize:      50,
		Compress:       true,
		Key:            "",
	}
}

func Load() Config {
	cfg := GetDefaultConfig()
	flag.StringVar(&cfg.ServerURL, "a", cfg.ServerURL, "server endpoint address in the format http://server:port")
	flag.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "fetching metrics frequency")
	flag.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "sending metrics frequency")
	flag.IntVar(&cfg.BatchSize, "b", cfg.BatchSize, "batch size")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "secret key for hash generation")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if !strings.HasPrefix(cfg.ServerURL, "http://") && !strings.HasPrefix(cfg.ServerURL, "https://") {
		cfg.ServerURL = "http://" + cfg.ServerURL
	}

	return cfg
}
