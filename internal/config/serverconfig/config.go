package serverconfig

import (
	"log"
	"flag"

	"github.com/caarlos0/env"
	//"github.com/joho/godotenv"
)

type Config struct {
	ServerURL string `env:"ADDRESS"`
	StoreInterval int `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore bool `env:"RESTORE"`
	DatabseDSN string `env:"DATABASE_DSN"`
}

func GetDefaultConfig() *Config{
	return &Config{
	ServerURL: "localhost:8080",
	StoreInterval: 300,
	FileStoragePath: "metrics.txt",
	Restore: true,
	}
}

func Load() *Config {
	cfg := GetDefaultConfig()
	// if err := godotenv.Load(); err != nil {
	// 	log.Println("Cannot load parameters from .env")
	// }
	
	flag.StringVar(&(cfg.ServerURL), "a", cfg.ServerURL, "Local listening interface in the format servername:port")
	flag.IntVar(&(cfg.StoreInterval), "i", cfg.StoreInterval, "The save to a file interval. If the value equals 0, it will enable a sync mode")
	flag.StringVar(&(cfg.FileStoragePath), "f", cfg.FileStoragePath, "The path to the file for saving metrics")
	flag.BoolVar(&(cfg.Restore), "r", cfg.Restore, "Do we need to restore metrics from the file during the start?")
	flag.StringVar(&cfg.DatabseDSN, "d", cfg.DatabseDSN, "DSN connectoin string")
	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}
