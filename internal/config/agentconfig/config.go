package agentconfig

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"

	"github.com/TheLuckymadman/metawatch/internal/config"
	"github.com/TheLuckymadman/metawatch/internal/model"
)

type Config struct {
	ServerURL      string         `env:"ADDRESS" json:"ADDRESS"`
	PollInterval   model.Duration `env:"POLL_INTERVAL" json:"POLL_INTERVAL"`
	ReportInterval model.Duration `env:"REPORT_INTERVAL" json:"REPORT_INTERVAL"`
	BatchSize      int            `env:"BATCH_SIZE"`
	Compress       bool           `env:"COMPRESS"`
	Key            string         `env:"KEY"`
	RateLimit      int            `env:"RATE_LIMIT"`
	LogMetrics     bool           `env:"LOG_METRICS"`
	CryptoKey      string         `env:"CRYPTO_KEY" json:"CRYPTO_KEY"`
	Config         string         `env:"CONFIG"`
}

func GetDefaultConfig() *Config {
	return &Config{
		ServerURL:      "http://127.0.0.1:8080",
		PollInterval:   model.Duration(2 * time.Second),
		ReportInterval: model.Duration(10 * time.Second),
		BatchSize:      50,
		Compress:       true,
		Key:            "",
		RateLimit:      3,
		LogMetrics:     false,
		CryptoKey:      "",
		Config:         "",
	}
}

func Load() *Config {
	cfg := GetDefaultConfig()

	var configFilePath string
	for i := 0; i < len(os.Args); i++ {
		if os.Args[i] == "-config" || os.Args[i] == "-c" {
			if i == len(os.Args)-1 {
				log.Fatalf("flag needs an argument: -%s", os.Args[i])
				break
			}
			configFilePath = os.Args[i+1]
			break
		} else if strings.HasPrefix(os.Args[i], "-config=") {
			configFilePath = strings.SplitN(os.Args[i], "=", 2)[1]
		} else if strings.HasPrefix(os.Args[i], "-c=") {
			configFilePath = strings.SplitN(os.Args[i], "=", 2)[1]
		}
	}

	if configFilePath != "" {
		if err := config.GetCfgFromFile(configFilePath, cfg); err != nil {
			log.Fatalf("%v", err)
		}
	}

	flag.StringVar(&cfg.ServerURL, "a", cfg.ServerURL, "server endpoint address in the format http://server:port")
	flag.Var(&cfg.PollInterval, "p", "fetching metrics frequency")
	flag.Var(&cfg.ReportInterval, "r", "sending metrics frequency")
	flag.IntVar(&cfg.BatchSize, "b", cfg.BatchSize, "batch size")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "secret key for hash generation")
	flag.IntVar(&cfg.RateLimit, "l", cfg.RateLimit, "max cucrurrent requests to server")
	flag.BoolVar(&cfg.LogMetrics, "lm", cfg.LogMetrics, "show metrics in the agent's log")
	flag.StringVar(&cfg.CryptoKey, "s", cfg.CryptoKey, "certificate path")
	flag.StringVar(&cfg.Config, "config", cfg.Config, "json config file path")
	flag.StringVar(&cfg.Config, "c", cfg.Config, "alias for -config")
	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		log.Fatalf("env parsing err: %v", err)
	}

	if !strings.HasPrefix(cfg.ServerURL, "http://") && !strings.HasPrefix(cfg.ServerURL, "https://") {
		cfg.ServerURL = "http://" + cfg.ServerURL
	}

	return cfg
}
