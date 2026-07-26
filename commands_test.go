package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitCmd(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	code := initCmd([]string{"test-agent"})
	if code != 0 {
		t.Fatalf("initCmd returned %d", code)
	}

	if _, err := os.Stat("test-agent.toml"); err != nil {
		t.Error("test-agent.toml should exist")
	}
}

func TestInitCmd_DefaultName(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	code := initCmd([]string{})
	if code != 0 {
		t.Fatalf("initCmd returned %d", code)
	}

	if _, err := os.Stat("my-agent.toml"); err != nil {
		t.Error("my-agent.toml should exist")
	}
}

func TestInitCmd_SaveError(t *testing.T) {
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)

	tmp := t.TempDir()
	readOnly := filepath.Join(tmp, "readonly")
	os.MkdirAll(readOnly, 0555)
	os.Chdir(readOnly)

	code := initCmd([]string{"test"})
	if code == 0 {
		t.Error("should return non-zero on save error")
	}
}

func TestValidateCmd_Valid(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "agent.toml")
	os.WriteFile(cfgPath, []byte(`
[agent]
name = "validator-test"
model = "claude"
`), 0644)

	code := validateCmd([]string{cfgPath})
	if code != 0 {
		t.Errorf("validateCmd returned %d, want 0", code)
	}
}

func TestValidateCmd_Invalid(t *testing.T) {
	code := validateCmd([]string{"/nonexistent/path.toml"})
	if code == 0 {
		t.Error("should return non-zero for missing file")
	}
}

func TestValidateCmd_MissingName(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "agent.toml")
	os.WriteFile(cfgPath, []byte(`
[agent]
model = "claude"
`), 0644)

	code := validateCmd([]string{cfgPath})
	if code == 0 {
		t.Error("should return non-zero for missing name")
	}
}

func TestListCmd_Empty(t *testing.T) {
	code := listCmd([]string{})
	if code != 0 {
		t.Errorf("listCmd returned %d, want 0", code)
	}
}

func TestLogsCmd_NoArgs(t *testing.T) {
	code := logsCmd([]string{})
	if code == 0 {
		t.Error("should return non-zero when no args")
	}
}

func TestLogsCmd_NotFound(t *testing.T) {
	code := logsCmd([]string{"nonexistent-agent-id"})
	if code == 0 {
		t.Error("should return non-zero for missing log file")
	}
}

func TestKillCmd_NoArgs(t *testing.T) {
	code := killCmd([]string{})
	if code == 0 {
		t.Error("should return non-zero when no args")
	}
}

func TestKillCmd_NotFound(t *testing.T) {
	code := killCmd([]string{"nonexistent-agent"})
	if code == 0 {
		t.Error("should return non-zero for nonexistent agent")
	}
}

func TestStatusCmd(t *testing.T) {
	code := statusCmd([]string{})
	if code != 0 {
		t.Errorf("statusCmd returned %d, want 0", code)
	}
}

func TestPrintHelp(t *testing.T) {
	printHelp()
}

func TestRememberCmd(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	original := os.Getenv("XDG_DATA_HOME")
	defer os.Setenv("XDG_DATA_HOME", original)
	os.Setenv("XDG_DATA_HOME", dir)

	code := rememberCmd([]string{"test-agent", "alice", "uses", "linux"})
	if code != 0 {
		t.Fatalf("rememberCmd returned %d", code)
	}
}

func TestRememberCmd_NoArgs(t *testing.T) {
	code := rememberCmd([]string{})
	if code == 0 {
		t.Error("should return non-zero with no args")
	}
}

func TestRememberCmd_ThreeArgs(t *testing.T) {
	dir := t.TempDir()
	original := os.Getenv("XDG_DATA_HOME")
	defer os.Setenv("XDG_DATA_HOME", original)
	os.Setenv("XDG_DATA_HOME", dir)

	code := rememberCmd([]string{"test-agent", "alice", "uses"})
	if code != 0 {
		t.Errorf("rememberCmd with 3 args returned %d", code)
	}
}

func TestRecallCmd_All(t *testing.T) {
	dir := t.TempDir()
	original := os.Getenv("XDG_DATA_HOME")
	defer os.Setenv("XDG_DATA_HOME", original)
	os.Setenv("XDG_DATA_HOME", dir)

	rememberCmd([]string{"test-agent", "alice", "uses", "linux"})
	code := recallCmd([]string{"test-agent"})
	if code != 0 {
		t.Errorf("recallCmd returned %d", code)
	}
}

func TestRecallCmd_WithQuery(t *testing.T) {
	dir := t.TempDir()
	original := os.Getenv("XDG_DATA_HOME")
	defer os.Setenv("XDG_DATA_HOME", original)
	os.Setenv("XDG_DATA_HOME", dir)

	rememberCmd([]string{"test-agent", "alice", "uses", "linux"})
	code := recallCmd([]string{"test-agent", "alice"})
	if code != 0 {
		t.Errorf("recallCmd returned %d", code)
	}
}

func TestRecallCmd_NoArgs(t *testing.T) {
	code := recallCmd([]string{})
	if code == 0 {
		t.Error("should return non-zero with no args")
	}
}

func TestForgetCmd(t *testing.T) {
	dir := t.TempDir()
	original := os.Getenv("XDG_DATA_HOME")
	defer os.Setenv("XDG_DATA_HOME", original)
	os.Setenv("XDG_DATA_HOME", dir)

	rememberCmd([]string{"test-agent", "alice", "uses", "linux"})
	code := forgetCmd([]string{"test-agent", "alice", "uses"})
	if code != 0 {
		t.Errorf("forgetCmd returned %d", code)
	}
}

func TestForgetCmd_NoArgs(t *testing.T) {
	code := forgetCmd([]string{})
	if code == 0 {
		t.Error("should return non-zero with no args")
	}
}

func TestValidateCmd_NoArgs(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	os.WriteFile("agent.toml", []byte(`
[agent]
name = "test"
model = "claude"
`), 0644)

	code := validateCmd([]string{})
	if code != 0 {
		t.Errorf("validateCmd with default path returned %d", code)
	}
}
