package serverconfig

import (
	"log"
	"flag"

	"github.com/caarlos0/env"
)

type Config struct {
	ServerURL string `env:"ADDRESS"`
}

func GetDefaultConfig() *Config{
	return &Config{
	ServerURL: "localhost:8080",
	}
}

func Load() *Config {
	cfg := GetDefaultConfig()
	flag.StringVar(&(cfg.ServerURL), "a", cfg.ServerURL, "local listening interface in the format servername:port")
	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}
