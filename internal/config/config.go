package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func GetCfgFromFile[T any](fileName string, cfg *T) error {
	if fileName != "" {
		f, err := os.Open(fileName)
		if err != nil {
			return fmt.Errorf("read config file error: %w", err)
		}
		defer f.Close()

		d := json.NewDecoder(f)
		err = d.Decode(cfg)
		if err != nil {
			return fmt.Errorf("json decoding config file error: %w", err)
		}
	}
	return nil
}
