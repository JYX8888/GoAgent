package core

import (
	"os"
	"strconv"
	"sync"
)

type Config struct {
	DefaultModel     string  `json:"default_model" env:"LLM_MODEL_ID"`
	DefaultProvider  string  `json:"default_provider" env:"LLM_PROVIDER"`
	Temperature      float64 `json:"temperature" env:"TEMPERATURE"`
	MaxTokens        *int    `json:"max_tokens" env:"MAX_TOKENS"`
	Debug            bool    `json:"debug" env:"DEBUG"`
	LogLevel         string  `json:"log_level" env:"LOG_LEVEL"`
	MaxHistoryLength int     `json:"max_history_length"`

	mu sync.RWMutex
}

func NewConfig() *Config {
	return &Config{
		DefaultModel:     "gpt-3.5-turbo",
		DefaultProvider:  "openai",
		Temperature:      0.7,
		MaxHistoryLength: 100,
		LogLevel:         "INFO",
		Debug:            false,
	}
}

func LoadConfigFromEnv() *Config {
	cfg := NewConfig()

	if v := os.Getenv("DEBUG"); v != "" {
		cfg.Debug = v == "true" || v == "1"
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := os.Getenv("TEMPERATURE"); v != "" {
		if temp, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.Temperature = temp
		}
	}
	if v := os.Getenv("MAX_TOKENS"); v != "" {
		if tokens, err := strconv.Atoi(v); err == nil {
			cfg.MaxTokens = &tokens
		}
	}
	if v := os.Getenv("LLM_MODEL_ID"); v != "" {
		cfg.DefaultModel = v
	}
	if v := os.Getenv("LLM_PROVIDER"); v != "" {
		cfg.DefaultProvider = v
	}

	return cfg
}

func (c *Config) SetMaxTokens(tokens int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.MaxTokens = &tokens
}

func (c *Config) GetMaxTokens() *int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.MaxTokens
}

func (c *Config) SetTemperature(temp float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Temperature = temp
}

func (c *Config) GetTemperature() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Temperature
}

func (c *Config) SetDebug(debug bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Debug = debug
}

func (c *Config) IsDebug() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Debug
}

func (c *Config) ToMap() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	m := map[string]interface{}{
		"default_model":      c.DefaultModel,
		"default_provider":   c.DefaultProvider,
		"temperature":        c.Temperature,
		"debug":              c.Debug,
		"log_level":          c.LogLevel,
		"max_history_length": c.MaxHistoryLength,
	}
	if c.MaxTokens != nil {
		m["max_tokens"] = *c.MaxTokens
	}
	return m
}
