package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "agent.toml")
	err := os.WriteFile(path, []byte(`
[agent]
name = "test-agent"
model = "claude-sonnet"
command = "echo"

[memory]
backend = "cairn"
path = "/tmp/test-memory.db"

[sandbox]
enabled = true
allow = ["./src"]
net = ["github.com"]

[mcp]
servers = ["filesystem", "git"]

[observe]
log = true
trace = false
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Agent.Name != "test-agent" {
		t.Errorf("name = %s, want test-agent", cfg.Agent.Name)
	}
	if cfg.Agent.Model != "claude-sonnet" {
		t.Errorf("model = %s", cfg.Agent.Model)
	}
	if cfg.Agent.Command != "echo" {
		t.Errorf("command = %s", cfg.Agent.Command)
	}
	if cfg.Memory.Backend != "cairn" {
		t.Errorf("backend = %s", cfg.Memory.Backend)
	}
	if !cfg.Sandbox.Enabled {
		t.Error("sandbox should be enabled")
	}
	if len(cfg.Sandbox.Allow) != 1 || cfg.Sandbox.Allow[0] != "./src" {
		t.Errorf("allow = %v", cfg.Sandbox.Allow)
	}
	if len(cfg.MCP.Servers) != 2 {
		t.Errorf("servers = %v", cfg.MCP.Servers)
	}
	if cfg.Observe.Log != true {
		t.Error("log should be true")
	}
	if cfg.Observe.Trace != false {
		t.Error("trace should be false")
	}
}

func TestLoad_MissingName(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "agent.toml")
	err := os.WriteFile(path, []byte(`
[agent]
model = "claude"
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Load(path)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/agent.toml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestDefault(t *testing.T) {
	cfg := Default("my-agent")
	if cfg.Agent.Name != "my-agent" {
		t.Errorf("name = %s", cfg.Agent.Name)
	}
	if cfg.Agent.Model != "claude-sonnet" {
		t.Errorf("model = %s", cfg.Agent.Model)
	}
	if !cfg.Sandbox.Enabled {
		t.Error("sandbox should be enabled by default")
	}
	if !cfg.Observe.Log {
		t.Error("log should be true by default")
	}
}

func TestSave(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "agent.toml")
	cfg := Default("test-save")
	if err := cfg.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Agent.Name != "test-save" {
		t.Errorf("name = %s", loaded.Agent.Name)
	}
}

func TestDataDir(t *testing.T) {
	d := DataDir()
	if d == "" {
		t.Error("DataDir should not be empty")
	}
}
