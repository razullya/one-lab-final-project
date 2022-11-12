package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DB_Host     string `json:"db_host"`
	DB_Port     string `json:"db_port"`
	DB_Username string `json:"db_username"`
	DB_Password string `json:"db_password"`
	DB_Name     string `json:"db_name"`
	DB_SSLMode  string `json:"db_ssl_mode"`
}

func InitConfig(path string) (*Config, error) {
	var cfg Config
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("init configs: open: %w", err)
	}
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("init configs: decode: %w", err)
	}
	return &cfg, nil
}
