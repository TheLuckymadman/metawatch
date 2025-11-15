package serverconfig

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"

	"github.com/TheLuckymadman/metawatch/internal/repository"
)

type Config struct {
	ServerURL       string                `env:"ADDRESS"`
	StoreInterval   int                   `env:"STORE_INTERVAL"`
	FileStoragePath string                `env:"FILE_STORAGE_PATH"`
	Restore         bool                  `env:"RESTORE"`
	DatabseDSN      string                `env:"DATABASE_DSN"`
	DBInitMode      repository.DBInitMode `env:"DB_INIT_MODE"`
}

func GetDefaultConfig() *Config {
	return &Config{
		ServerURL:       "localhost:8080",
		StoreInterval:   300,
		FileStoragePath: "metrics.txt",
		Restore:         true,
	}
}

func Load() *Config {
	cfg := GetDefaultConfig()
	if err := godotenv.Load(); err != nil {
		log.Println("Cannot load parameters from .env")
	}

	var dbInitMode string
	flag.StringVar(&(cfg.ServerURL), "a", cfg.ServerURL, "Local listening interface in the format servername:port")
	flag.IntVar(&(cfg.StoreInterval), "i", cfg.StoreInterval, "The save to a file interval. If the value equals 0, it will enable a sync mode")
	flag.StringVar(&(cfg.FileStoragePath), "f", cfg.FileStoragePath, "The path to the file for saving metrics")
	flag.BoolVar(&(cfg.Restore), "r", cfg.Restore, "Do we need to restore metrics from the file during the start?")
	flag.StringVar(&cfg.DatabseDSN, "d", cfg.DatabseDSN, "DSN connectoin string")
	flag.StringVar(&dbInitMode, "m", "internal", "DB init mode, use external, internal, reset")
	flag.Parse()

	switch dbInitMode {
	case "external", "0":
		cfg.DBInitMode = repository.ManagedExternally
	case "internal", "1":
		cfg.DBInitMode = repository.ManagedInternally
	case "reset", "forced", "2":
		cfg.DBInitMode = repository.ManagedInternallyForced
	default:
		cfg.DBInitMode = repository.ManagedExternally
	}

	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}
