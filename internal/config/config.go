package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration loaded from config.yaml.
type Config struct {
	Spoolman SpoolmanConfig  `yaml:"spoolman"`
	Printers []PrinterConfig `yaml:"printers"`
}

// SpoolmanConfig holds the Spoolman server connection details.
type SpoolmanConfig struct {
	Address string `yaml:"address"`
}

// PrinterConfig holds the configuration for a single Bambu Lab printer.
type PrinterConfig struct {
	Name       string `yaml:"name"`
	IP         string `yaml:"ip"`
	Serial     string `yaml:"serial"`
	AccessCode string `yaml:"access_code"`
}

// LoadConfig reads and parses a YAML config file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Spoolman.Address == "" {
		return fmt.Errorf("spoolman.address is required")
	}
	if len(c.Printers) == 0 {
		return fmt.Errorf("at least one printer must be configured")
	}
	for i, p := range c.Printers {
		if p.IP == "" {
			return fmt.Errorf("printers[%d].ip is required", i)
		}
		if p.Serial == "" {
			return fmt.Errorf("printers[%d].serial is required", i)
		}
		if p.AccessCode == "" {
			return fmt.Errorf("printers[%d].access_code is required", i)
		}
		if p.Name == "" {
			c.Printers[i].Name = p.Serial
		}
	}
	return nil
}
