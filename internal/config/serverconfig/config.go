package serverconfig

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"

	"github.com/TheLuckymadman/metawatch/internal/config"
	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/TheLuckymadman/metawatch/internal/repository"
)

type Config struct {
	ServerURL       string                `env:"ADDRESS" json:"ADDRESS"`
	StoreInterval   model.Duration        `env:"STORE_INTERVAL" json:"STORE_INTERVAL"`
	FileStoragePath string                `env:"FILE_STORAGE_PATH" json:"STORE_FILE"`
	Restore         bool                  `env:"RESTORE" json:"RESTORE"`
	DatabaseDSN     string                `env:"DATABASE_DSN" json:"DATABASE_DSN"`
	DBInitMode      repository.DBInitMode `env:"DB_INIT_MODE"`
	Key             string                `env:"KEY"`
	AuditFile       string                `env:"AUDIT_FILE"`
	AuditURL        string                `env:"AUDIT_URL"`
	CryptoKey       string                `env:"CRYPTO_KEY" json:"CRYPTO_KEY"`
	Config          string                `env:"CONFIG"`
}

func GetDefaultConfig() *Config {
	return &Config{
		ServerURL:       "localhost:8080",
		StoreInterval:   model.Duration(300 * time.Second),
		FileStoragePath: "metrics.txt",
		Restore:         false,
		Key:             "",
		AuditFile:       "",
		AuditURL:        "",
		CryptoKey:       "",
		Config:          "",
	}
}

func Load() *Config {
	cfg := GetDefaultConfig()
	if err := godotenv.Load(); err != nil {
		log.Println("Cannot load parameters from .env")
	}

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

	var dbInitMode string
	flag.StringVar(&cfg.ServerURL, "a", cfg.ServerURL, "Local listening interface in the format servername:port")
	flag.Var(&cfg.StoreInterval, "i", "The save to a file interval. If the value equals 0, it will enable a sync mode")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "The path to the file for saving metrics")
	flag.BoolVar(&cfg.Restore, "r", cfg.Restore, "Do we need to restore metrics from the file during the start?")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "DSN connectoin string")
	flag.StringVar(&dbInitMode, "m", "internal", "DB init mode, use external, internal, reset")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "Secret key for hash generation")
	flag.StringVar(&cfg.AuditFile, "audit-file", cfg.AuditFile, "Audit file path")
	flag.StringVar(&cfg.AuditURL, "audit-url", cfg.AuditURL, "Audit server url")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "private key path")
	flag.StringVar(&cfg.Config, "config", cfg.Config, "json config file path")
	flag.StringVar(&cfg.Config, "c", cfg.Config, "alias for -config")
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
