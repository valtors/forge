package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Agent   AgentConfig   `toml:"agent"`
	Memory  MemoryConfig  `toml:"memory"`
	Sandbox SandboxConfig `toml:"sandbox"`
	MCP     MCPConfig     `toml:"mcp"`
	Observe ObserveConfig `toml:"observe"`
}

type AgentConfig struct {
	Name     string `toml:"name"`
	Model    string `toml:"model"`
	Provider string `toml:"provider"`
	Command  string `toml:"command"`
}

type MemoryConfig struct {
	Backend string `toml:"backend"`
	Path    string `toml:"path"`
}

type SandboxConfig struct {
	Enabled bool     `toml:"enabled"`
	Allow   []string `toml:"allow"`
	Net     []string `toml:"net"`
}

type MCPConfig struct {
	Servers []string `toml:"servers"`
}

type ObserveConfig struct {
	Log   bool `toml:"log"`
	Trace bool `toml:"trace"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = "agent.toml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Agent.Name == "" {
		return nil, fmt.Errorf("agent.name is required")
	}

	return &cfg, nil
}

func Default(name string) *Config {
	return &Config{
		Agent: AgentConfig{
			Name:    name,
			Model:   "claude-sonnet",
			Command: "claude",
		},
		Memory: MemoryConfig{
			Backend: "cairn",
			Path:    filepath.Join(dataDir(), "forge", name, "memory.db"),
		},
		Sandbox: SandboxConfig{
			Enabled: true,
			Allow:   []string{"./src"},
			Net:     []string{},
		},
		MCP: MCPConfig{
			Servers: []string{},
		},
		Observe: ObserveConfig{
			Log:   true,
			Trace: true,
		},
	}
}

func (c *Config) Save(path string) error {
	if path == "" {
		path = "agent.toml"
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	return enc.Encode(c)
}

func dataDir() string {
	if d := os.Getenv("XDG_DATA_HOME"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share")
}

func DataDir() string {
	return dataDir()
}
