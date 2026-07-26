package sandbox

import (
	"testing"
)

func TestSanitizeEnv(t *testing.T) {
	c := New(true, []string{}, []string{})
	env := []string{
		"PATH=/usr/bin",
		"GITHUB_TOKEN=ghp_secret",
		"OPENAI_API_KEY=sk_secret",
		"HOME=/home/user",
		"DATABASE_URL=postgres://localhost",
		"MY_VAR=value",
	}
	clean := c.SanitizeEnv(env)
	if len(clean) != 3 {
		t.Fatalf("got %d env vars, want 3: %v", len(clean), clean)
	}
	for _, e := range clean {
		if e == "GITHUB_TOKEN=ghp_secret" {
			t.Error("GITHUB_TOKEN should be stripped")
		}
	}
}

func TestSanitizeEnv_Disabled(t *testing.T) {
	c := New(false, []string{}, []string{})
	env := []string{"SECRET_TOKEN=xxx", "PATH=/usr/bin"}
	clean := c.SanitizeEnv(env)
	if len(clean) != 2 {
		t.Errorf("got %d, want 2 (disabled should not strip)", len(clean))
	}
}

func TestAllowedPath(t *testing.T) {
	c := New(true, []string{"/home/user/project"}, []string{})

	if !c.AllowedPath("/home/user/project/src/main.go") {
		t.Error("project path should be allowed")
	}
	if !c.AllowedPath("/home/user/project") {
		t.Error("project root should be allowed")
	}
}

func TestBlockedPath(t *testing.T) {
	c := New(true, []string{"/home/user/project"}, []string{})

	if c.AllowedPath("/home/user/.ssh/id_rsa") {
		t.Error(".ssh should be blocked")
	}
	if c.AllowedPath("/home/user/.aws/credentials") {
		t.Error(".aws should be blocked")
	}
	if c.AllowedPath("/home/user/.env") {
		t.Error(".env should be blocked")
	}
}

func TestAllowedHost(t *testing.T) {
	c := New(true, []string{}, []string{"github.com", "*.npmjs.org"})

	if !c.AllowedHost("github.com") {
		t.Error("github.com should be allowed")
	}
	if !c.AllowedHost("registry.npmjs.org") {
		t.Error("*.npmjs.org should match registry.npmjs.org")
	}
	if c.AllowedHost("evil.com") {
		t.Error("evil.com should be blocked")
	}
}

func TestAllowedHost_Disabled(t *testing.T) {
	c := New(true, []string{}, []string{})

	if !c.AllowedHost("anything.com") {
		t.Error("empty net list should allow all")
	}
}

func TestSetup(t *testing.T) {
	c := New(true, []string{}, []string{})
	if err := c.Setup(); err != nil {
		t.Fatal(err)
	}
	if c.OverlayDir() == "" {
		t.Error("overlay dir should be set")
	}
	c.Cleanup()
}

func TestSetup_Disabled(t *testing.T) {
	c := New(false, []string{}, []string{})
	if err := c.Setup(); err != nil {
		t.Fatal(err)
	}
	if c.OverlayDir() != "" {
		t.Error("overlay dir should be empty when disabled")
	}
}

func TestBlockedPaths(t *testing.T) {
	paths := BlockedPaths()
	if len(paths) == 0 {
		t.Error("should have blocked paths")
	}
	found := false
	for _, p := range paths {
		if p == ".ssh" {
			found = true
		}
	}
	if !found {
		t.Error(".ssh should be in blocked paths")
	}
}

func TestSecretPatterns(t *testing.T) {
	patterns := SecretPatterns()
	if len(patterns) == 0 {
		t.Error("should have secret patterns")
	}
}

func TestIsSecret(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"GITHUB_TOKEN", true},
		{"OPENAI_API_KEY", true},
		{"PASSWORD", true},
		{"AWS_SECRET_ACCESS_KEY", true},
		{"PATH", false},
		{"HOME", false},
		{"MY_VAR", false},
		{"DATABASE_URL", true},
	}
	for _, tt := range tests {
		if got := isSecret(tt.key); got != tt.want {
			t.Errorf("isSecret(%s) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestAllowedPath_Disabled(t *testing.T) {
	c := New(false, []string{}, []string{})
	if !c.AllowedPath("/anywhere/random") {
		t.Error("disabled sandbox should allow all paths")
	}
}

func TestSanitizeEnv_Empty(t *testing.T) {
	c := New(true, []string{}, []string{})
	clean := c.SanitizeEnv([]string{})
	if len(clean) != 0 {
		t.Errorf("got %d, want 0", len(clean))
	}
}

func TestSanitizeEnv_MultipleSecrets(t *testing.T) {
	c := New(true, []string{}, []string{})
	env := []string{
		"PATH=/usr/bin",
		"GITHUB_TOKEN=xxx",
		"OPENAI_API_KEY=xxx",
		"ANTHROPIC_API_KEY=xxx",
		"AWS_SECRET_ACCESS_KEY=xxx",
		"AZURE_CLIENT_SECRET=xxx",
		"DATABASE_URL=xxx",
		"PRIVATE_KEY=xxx",
		"HOME=/home/user",
	}
	clean := c.SanitizeEnv(env)
	if len(clean) != 2 {
		t.Errorf("got %d, want 2 (PATH and HOME only)", len(clean))
	}
}

func TestAllowedHost_Wildcard(t *testing.T) {
	c := New(true, []string{}, []string{"*.github.com"})

	if !c.AllowedHost("api.github.com") {
		t.Error("api.github.com should match *.github.com")
	}
	if !c.AllowedHost("raw.github.com") {
		t.Error("raw.github.com should match *.github.com")
	}
	if c.AllowedHost("github.com") {
		t.Error("github.com should NOT match *.github.com")
	}
}

func TestAllowedHost_ExactMatch(t *testing.T) {
	c := New(true, []string{}, []string{"github.com"})

	if !c.AllowedHost("github.com") {
		t.Error("exact match should work")
	}
	if c.AllowedHost("api.github.com") {
		t.Error("should not match wildcard when only exact specified")
	}
}

func TestIsSecret_CaseInsensitive(t *testing.T) {
	if !isSecret("github_token") {
		t.Error("lowercase github_token should be secret")
	}
	if !isSecret("GITHUB_TOKEN") {
		t.Error("uppercase GITHUB_TOKEN should be secret")
	}
	if !isSecret("Github_Token") {
		t.Error("mixed case should be secret")
	}
}

func TestIsSecret_NotSecret(t *testing.T) {
	if isSecret("EDITOR") {
		t.Error("EDITOR should not be secret")
	}
	if isSecret("SHELL") {
		t.Error("SHELL should not be secret")
	}
	if isSecret("TERM") {
		t.Error("TERM should not be secret")
	}
}
