package main

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Model       string `yaml:"model"`
	LibPath     string `yaml:"lib_path"`
	DBPath      string `yaml:"db_path"`
	ContextSize int    `yaml:"context_size"`
	BatchSize   int    `yaml:"batch_size"`
	Verbose     bool   `yaml:"verbose"`
	Server      struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
}

func DefaultConfig() *Config {
	return &Config{
		Model:       "",
		LibPath:     "",
		DBPath:      "rag.db",
		ContextSize: 512,
		BatchSize:   512,
		Verbose:     false,
		Server: struct {
			Port string `yaml:"port"`
		}{Port: "8080"},
	}
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path != "" {
		if data, err := os.ReadFile(path); err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, err
			}
		}
	}

	cfg.applyEnvOverrides()
	return cfg, nil
}

func (c *Config) applyEnvOverrides() {
	if v := os.Getenv("YDRAG_MODEL"); v != "" {
		c.Model = v
	}
	if v := os.Getenv("YZMA_LIB"); v != "" {
		c.LibPath = v
	}
	if v := os.Getenv("YDRAG_DB_PATH"); v != "" {
		c.DBPath = v
	}
	if v := os.Getenv("YDRAG_CONTEXT_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.ContextSize = n
		}
	}
	if v := os.Getenv("YDRAG_BATCH_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.BatchSize = n
		}
	}
	if v := os.Getenv("YDRAG_VERBOSE"); v != "" {
		c.Verbose = v == "true" || v == "1"
	}
	if v := os.Getenv("YDRAG_SERVER_PORT"); v != "" {
		c.Server.Port = v
	}
}
