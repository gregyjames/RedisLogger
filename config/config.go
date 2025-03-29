package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	ListenAddr string `json:"listen_addr"`
	RedisAddr  string `json:"redis_addr"`
}

func Load(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %v", err)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("error decoding config: %v", err)
	}

	return &config, nil
} 