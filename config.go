package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	configPath = "config.yml"
)

type config struct {
	Listen         string        `yaml:"listen"`
	Whitelist      []string      `yaml:"whitelist"`
	UpdateInterval time.Duration `yaml:"refresh_interval"`
	Subnet         string        `yaml:"subnet"`
	SubnetMask     int           `yaml:"subnet_mask"`
}

var cfg *config

func loadConfig() error {
	f, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}

	return nil
}
