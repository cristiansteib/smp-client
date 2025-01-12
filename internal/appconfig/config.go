package appconfig

import (
	"encoding/json"
	"fmt"
	"os"
)

type AppConfig struct {
	InfluxURL    string `json:"influx_url"`
	InfluxOrg    string `json:"influx_org"`
	InfluxBucket string `json:"influx_bucket"`
	InfluxToken  string `json:"influx_token"`
}

func (c *AppConfig) validateConfig() error {
	if c.InfluxURL == "" {
		return fmt.Errorf("InfluxURL is empty")
	}
	if c.InfluxToken == "" {
		return fmt.Errorf("InfluxToken is empty")
	}
	return nil
}

func LoadConfig(filePath string) (*AppConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config AppConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}
	if err = config.validateConfig(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}
	return &config, nil
}
