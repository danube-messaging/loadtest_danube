package config

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// LoadFile reads a YAML config from disk and unmarshals it into Config.
func LoadFile(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	return &cfg, nil
}
