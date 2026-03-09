package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AlwaysExecCLI  bool `yaml:"always_exec_cli,omitempty"`
	AlwaysWriteFile bool `yaml:"always_write_file,omitempty"`
}

func DefaultPath() string {
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "mcp2win", "config.yaml")
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultPath()
	}
	cfg := &Config{}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Save(path string, cfg *Config) error {
	if path == "" {
		path = DefaultPath()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
