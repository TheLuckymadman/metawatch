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
	DatabaseDSN      string                `env:"DATABASE_DSN"`
	DBInitMode      repository.DBInitMode `env:"DB_INIT_MODE"`
	Key             string                `env:"KEY"`
	AuditFile       string                `env:"AUDIT_FILE"`
	AuditURL        string                `env:"AUDIT_URL"`
}

func GetDefaultConfig() *Config {
	return &Config{
		ServerURL:       "localhost:8080",
		StoreInterval:   300,
		FileStoragePath: "metrics.txt",
		Restore:         false,
		Key:             "",
		AuditFile:       "",
		AuditURL:        "",
	}
}

func Load() *Config {
	cfg := GetDefaultConfig()
	if err := godotenv.Load(); err != nil {
		log.Println("Cannot load parameters from .env")
	}

	var dbInitMode string
	flag.StringVar(&cfg.ServerURL, "a", cfg.ServerURL, "Local listening interface in the format servername:port")
	flag.IntVar(&cfg.StoreInterval, "i", cfg.StoreInterval, "The save to a file interval. If the value equals 0, it will enable a sync mode")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "The path to the file for saving metrics")
	flag.BoolVar(&cfg.Restore, "r", cfg.Restore, "Do we need to restore metrics from the file during the start?")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "DSN connectoin string")
	flag.StringVar(&dbInitMode, "m", "internal", "DB init mode, use external, internal, reset")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "Secret key for hash generation")
	flag.StringVar(&cfg.AuditFile, "audit-file", cfg.AuditFile, "Audit file path")
	flag.StringVar(&cfg.AuditURL, "audit-url", cfg.AuditURL, "Audit server url")
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
