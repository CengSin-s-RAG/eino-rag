// Package config loads the explicit runtime configuration.
package config

import "github.com/ilyakaznacheev/cleanenv"

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"DB"`
}

type ExecConfig struct {
	AllowedCommands []string `yaml:"allowedCommands"`
	Timeout         int      `yaml:"timeout"`
}

type ProviderConfig struct {
	Type        string  `yaml:"type"`
	BaseURL     string  `yaml:"baseURL"`
	APIKeyEnv   string  `yaml:"apiKeyEnv"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
}

type ProvidersConfig struct {
	Primary  ProviderConfig   `yaml:"primary"`
	Fallback []ProviderConfig `yaml:"fallback"`
}

type AgentConfig struct {
	HistoryLimit int `yaml:"historyLimit"`
	MaxSteps     int `yaml:"maxSteps"`
}

type MCPConfig struct {
	Name     string `yaml:"name"`
	Endpoint string `yaml:"endpoint"`
}

type Config struct {
	MCP        []MCPConfig     `yaml:"mcp"`
	Redis      *RedisConfig    `yaml:"redis"`
	SkillsPath string          `yaml:"skillsPath"`
	Exec       *ExecConfig     `yaml:"exec"`
	Providers  ProvidersConfig `yaml:"providers"`
	Agent      AgentConfig     `yaml:"agent"`
}

func Load(path string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
