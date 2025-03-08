package config

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

// Config holds the application configuration
type Config struct {
	StoreID      int           `yaml:"store_id"`
	CheckInterval time.Duration `yaml:"check_interval"`
	Products     []Product     `yaml:"products"`
	SeleniumOpts SeleniumOptions `yaml:"selenium"`
}

// Product represents a product to check inventory for
type Product struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// SeleniumOptions contains settings for the Selenium browser
type SeleniumOptions struct {
	ChromeDriverPath string `yaml:"chrome_driver_path"`
	Headless         bool   `yaml:"headless"`
	Timeout          time.Duration `yaml:"timeout"`
}

// Load loads the configuration from a file
func Load(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	// Set defaults
	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = 5 * time.Minute
	}
	if cfg.SeleniumOpts.Timeout == 0 {
		cfg.SeleniumOpts.Timeout = 30 * time.Second
	}

	return &cfg, nil
}
