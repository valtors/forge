package agent

import (
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager(t.TempDir())
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if len(m.List()) != 0 {
		t.Error("new manager should have no agents")
	}
}

func TestStartAndStop(t *testing.T) {
	m := NewManager(t.TempDir())
	logPath := filepath.Join(t.TempDir(), "test.log")

	err := m.Start("agent-1", "test", "sleep", []string{"10"}, logPath)
	if err != nil {
		t.Fatal(err)
	}

	a, exists := m.Get("agent-1")
	if !exists {
		t.Fatal("agent not found after start")
	}
	if a.Status() != StatusRunning {
		t.Errorf("status = %s, want running", a.Status())
	}
	if a.PID() == 0 {
		t.Error("PID should not be 0")
	}

	err = m.Stop("agent-1")
	if err != nil {
		t.Fatal(err)
	}
	if a.Status() != StatusStopped {
		t.Errorf("status = %s, want stopped", a.Status())
	}
}

func TestStart_Duplicate(t *testing.T) {
	m := NewManager(t.TempDir())
	logPath := filepath.Join(t.TempDir(), "test.log")

	m.Start("agent-1", "test", "sleep", []string{"10"}, logPath)
	err := m.Start("agent-1", "test", "sleep", []string{"10"}, logPath)
	if err == nil {
		t.Error("should error on duplicate start")
	}
	m.Stop("agent-1")
}

func TestStop_NotFound(t *testing.T) {
	m := NewManager(t.TempDir())
	err := m.Stop("nonexistent")
	if err == nil {
		t.Error("should error on nonexistent agent")
	}
}

func TestList(t *testing.T) {
	m := NewManager(t.TempDir())
	logDir := t.TempDir()

	m.Start("a1", "test1", "sleep", []string{"10"}, filepath.Join(logDir, "a1.log"))
	m.Start("a2", "test2", "sleep", []string{"10"}, filepath.Join(logDir, "a2.log"))

	agents := m.List()
	if len(agents) != 2 {
		t.Errorf("got %d agents, want 2", len(agents))
	}

	m.Stop("a1")
	m.Stop("a2")
}

func TestGet_NotFound(t *testing.T) {
	m := NewManager(t.TempDir())
	_, exists := m.Get("nonexistent")
	if exists {
		t.Error("should not find nonexistent agent")
	}
}

func TestUptime(t *testing.T) {
	m := NewManager(t.TempDir())
	logPath := filepath.Join(t.TempDir(), "test.log")
	m.Start("a1", "test", "sleep", []string{"10"}, logPath)

	a, _ := m.Get("a1")
	up := a.Uptime()
	if up <= 0 {
		t.Error("uptime should be positive")
	}

	m.Stop("a1")
}

func TestGetPID(t *testing.T) {
	m := NewManager(t.TempDir())
	logPath := filepath.Join(t.TempDir(), "test.log")
	m.Start("a1", "test", "sleep", []string{"10"}, logPath)

	pid := m.GetPID("a1")
	if pid == 0 {
		t.Error("PID should not be 0 for running agent")
	}

	m.Stop("a1")
}

func TestGetPID_NotFound(t *testing.T) {
	m := NewManager(t.TempDir())
	pid := m.GetPID("nonexistent")
	if pid != 0 {
		t.Error("PID should be 0 for nonexistent agent")
	}
}

func TestStart_InvalidCommand(t *testing.T) {
	m := NewManager(t.TempDir())
	logPath := filepath.Join(t.TempDir(), "test.log")
	err := m.Start("a1", "test", "nonexistent-command-xyz", []string{}, logPath)
	if err == nil {
		t.Error("should error on invalid command")
	}
}

func TestStart_Crashed(t *testing.T) {
	m := NewManager(t.TempDir())
	logPath := filepath.Join(t.TempDir(), "test.log")

	err := m.Start("a1", "test", "false", []string{}, logPath)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)
	a, _ := m.Get("a1")
	if a.Status() != StatusCrashed {
		t.Errorf("status = %s, want crashed", a.Status())
	}
}
