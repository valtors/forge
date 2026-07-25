package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var blockedPaths = []string{
	".ssh", ".aws", ".gnupg", ".docker", ".kube",
	".config/gcloud", ".config/gh", ".npmrc",
	".pypirc", ".netrc", ".env", ".gitconfig",
}

var secretPatterns = []string{
	"TOKEN", "SECRET", "PASSWORD", "CREDENTIAL",
	"API_KEY", "AUTH", "AWS_", "AZURE_",
	"OPENAI", "ANTHROPIC", "CLAUDE", "STRIPE",
	"RESEND", "MAILGUN", "SENDGRID", "DATABASE_URL",
	"DSN", "PRIVATE_KEY", "NPM_TOKEN", "GITHUB_TOKEN",
	"GH_PAT", "GOOGLE_API",
}

type Config struct {
	Enabled    bool
	Allow      []string
	Net        []string
	overlayDir string
}

func New(enabled bool, allow, net []string) *Config {
	return &Config{
		Enabled: enabled,
		Allow:   allow,
		Net:     net,
	}
}

func (c *Config) Setup() error {
	if !c.Enabled {
		return nil
	}

	dir, err := os.MkdirTemp("", "forge-sandbox-*")
	if err != nil {
		return fmt.Errorf("create overlay: %w", err)
	}
	c.overlayDir = dir
	return nil
}

func (c *Config) Cleanup() {
	if c.overlayDir != "" {
		_ = os.RemoveAll(c.overlayDir)
	}
}

func (c *Config) OverlayDir() string {
	return c.overlayDir
}

func (c *Config) SanitizeEnv(env []string) []string {
	if !c.Enabled {
		return env
	}

	var clean []string
	for _, e := range env {
		parts := strings.SplitN(e, "=", 1)
		if len(parts) < 1 {
			continue
		}
		key := parts[0]
		if isSecret(key) {
			continue
		}
		clean = append(clean, e)
	}
	return clean
}

func (c *Config) AllowedPath(path string) bool {
	if !c.Enabled {
		return true
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	for _, blocked := range blockedPaths {
		if strings.Contains(abs, "/"+blocked+"/") || strings.HasSuffix(abs, "/"+blocked) {
			return false
		}
	}

	for _, allowed := range c.Allow {
		absAllowed, err := filepath.Abs(allowed)
		if err != nil {
			continue
		}
		if strings.HasPrefix(abs, absAllowed) {
			return true
		}
	}

	return false
}

func (c *Config) AllowedHost(host string) bool {
	if !c.Enabled || len(c.Net) == 0 {
		return true
	}

	for _, allowed := range c.Net {
		if matchHost(allowed, host) {
			return true
		}
	}
	return false
}

func matchHost(pattern, host string) bool {
	if pattern == host {
		return true
	}
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:]
		return strings.HasSuffix(host, suffix)
	}
	return false
}

func isSecret(key string) bool {
	upper := strings.ToUpper(key)
	for _, pattern := range secretPatterns {
		if strings.Contains(upper, pattern) {
			return true
		}
	}
	return false
}

func BlockedPaths() []string {
	return blockedPaths
}

func SecretPatterns() []string {
	return secretPatterns
}
