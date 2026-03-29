package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all configuration values.
type Config struct {
	APIKey     string `json:"api_key"`
	BaseURL    string `json:"base_url"`
	Model      string `json:"model"`
	TargetLang string `json:"target_lang"`
	Verbose    bool   `json:"-"`
}

// Defaults.
const (
	DefaultBaseURL    = "https://api.openai.com/v1"
	DefaultModel      = "gpt-4o-mini"
	DefaultTargetLang = "zh"
)

// configPath returns the path to ~/.trans.json.
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home dir: %w", err)
	}
	return filepath.Join(home, ".trans.json"), nil
}

// loadFile reads ~/.trans.json if it exists.
func loadFile() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return &cfg, nil
}

// applyEnv overlays environment variables onto the config.
func applyEnv(cfg *Config) {
	if v := os.Getenv("OPENAI_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("OPENAI_BASE_URL"); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv("TRANS_MODEL"); v != "" {
		cfg.Model = v
	}
	if v := os.Getenv("TRANS_TARGET_LANG"); v != "" {
		cfg.TargetLang = v
	}
}

// applyDefaults fills in zero values with defaults.
func applyDefaults(cfg *Config) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	if cfg.Model == "" {
		cfg.Model = DefaultModel
	}
	if cfg.TargetLang == "" {
		cfg.TargetLang = DefaultTargetLang
	}
}

// Load returns the final config: file → env → defaults.
// CLI flags are applied separately via the Setters below.
func Load() (*Config, error) {
	cfg, err := loadFile()
	if err != nil {
		return nil, err
	}
	applyEnv(cfg)
	applyDefaults(cfg)
	return cfg, nil
}

// SetModel overrides the model (from CLI flag).
func (c *Config) SetModel(m string) {
	if m != "" {
		c.Model = m
	}
}

// SetTargetLang overrides the target language (from CLI flag).
func (c *Config) SetTargetLang(lang string) {
	if lang != "" {
		c.TargetLang = lang
	}
}

// SetVerbose enables verbose output.
func (c *Config) SetVerbose(v bool) {
	c.Verbose = v
}

// Validate checks that required fields are present.
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key required: set OPENAI_API_KEY or add api_key to ~/.trans.json")
	}
	return nil
}
