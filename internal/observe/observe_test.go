package observe

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	tmp := t.TempDir()
	l, err := NewLogger(tmp, "test-agent", true, true)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
}

func TestLog(t *testing.T) {
	tmp := t.TempDir()
	l, err := NewLogger(tmp, "test-agent", true, true)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	l.Log(Entry{
		Timestamp: time.Now(),
		AgentID:   "test-agent",
		Tool:      "read_file",
		Input:     "/tmp/test.txt",
		Output:    "hello world",
		Duration:  5 * time.Millisecond,
	})

	history := l.History(10)
	if len(history) != 1 {
		t.Fatalf("got %d entries, want 1", len(history))
	}
	if history[0].Tool != "read_file" {
		t.Errorf("tool = %s", history[0].Tool)
	}
}

func TestLog_Disabled(t *testing.T) {
	tmp := t.TempDir()
	l, err := NewLogger(tmp, "test-agent", false, false)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	l.Log(Entry{
		Timestamp: time.Now(),
		AgentID:   "test-agent",
		Tool:      "read_file",
	})

	if history := l.History(10); len(history) != 0 {
		t.Errorf("got %d entries, want 0 (disabled)", len(history))
	}
}

func TestHistory_Limit(t *testing.T) {
	tmp := t.TempDir()
	l, err := NewLogger(tmp, "test-agent", true, false)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	for i := 0; i < 20; i++ {
		l.Log(Entry{
			Timestamp: time.Now(),
			AgentID:   "test-agent",
			Tool:      "tool",
		})
	}

	if history := l.History(5); len(history) != 5 {
		t.Errorf("got %d, want 5", len(history))
	}
}

func TestSearch(t *testing.T) {
	tmp := t.TempDir()
	l, err := NewLogger(tmp, "test-agent", true, false)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	l.Log(Entry{Timestamp: time.Now(), AgentID: "a", Tool: "read_file", Input: "test.txt"})
	l.Log(Entry{Timestamp: time.Now(), AgentID: "a", Tool: "write_file", Input: "output.log"})
	l.Log(Entry{Timestamp: time.Now(), AgentID: "a", Tool: "delete_file", Input: "temp.bak"})

	results := l.Search("read", 10)
	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
	if results[0].Tool != "read_file" {
		t.Errorf("tool = %s", results[0].Tool)
	}
}

func TestStats(t *testing.T) {
	tmp := t.TempDir()
	l, err := NewLogger(tmp, "test-agent", true, false)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	l.Log(Entry{Timestamp: time.Now(), AgentID: "a", Tool: "read_file"})
	l.Log(Entry{Timestamp: time.Now(), AgentID: "a", Tool: "read_file"})
	l.Log(Entry{Timestamp: time.Now(), AgentID: "a", Tool: "write_file"})

	stats := l.Stats()
	if stats["read_file"] != 2 {
		t.Errorf("read_file = %d, want 2", stats["read_file"])
	}
	if stats["write_file"] != 1 {
		t.Errorf("write_file = %d, want 1", stats["write_file"])
	}
}

func TestLogFileWritten(t *testing.T) {
	tmp := t.TempDir()
	l, err := NewLogger(tmp, "test-agent", true, true)
	if err != nil {
		t.Fatal(err)
	}

	l.Log(Entry{
		Timestamp: time.Now(),
		AgentID:   "test-agent",
		Tool:      "test_tool",
		Input:     "test_input",
		Output:    "test_output",
	})
	l.Close()

	data, err := os.ReadFile(filepath.Join(tmp, "test-agent.log"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("log file should not be empty")
	}
}

func TestNewLogger_InvalidDir(t *testing.T) {
	_, err := NewLogger("/nonexistent/path/that/does/not/exist", "test", true, true)
	if err == nil {
		t.Error("should error on invalid dir")
	}
}

func TestLog_Error(t *testing.T) {
	tmp := t.TempDir()
	l, _ := NewLogger(tmp, "test", true, false)
	defer l.Close()

	l.Log(Entry{
		Timestamp: time.Now(),
		AgentID:   "test",
		Tool:      "failing_tool",
		Error:     "permission denied",
	})

	history := l.History(10)
	if len(history) != 1 {
		t.Fatalf("got %d, want 1", len(history))
	}
	if history[0].Error != "permission denied" {
		t.Errorf("error = %s", history[0].Error)
	}
}

func TestSearch_NoMatch(t *testing.T) {
	tmp := t.TempDir()
	l, _ := NewLogger(tmp, "test", true, false)
	defer l.Close()

	l.Log(Entry{Timestamp: time.Now(), Tool: "read_file", Input: "test"})

	results := l.Search("nonexistent", 10)
	if len(results) != 0 {
		t.Errorf("got %d, want 0", len(results))
	}
}

func TestSearch_Limit(t *testing.T) {
	tmp := t.TempDir()
	l, _ := NewLogger(tmp, "test", true, false)
	defer l.Close()

	for i := 0; i < 10; i++ {
		l.Log(Entry{Timestamp: time.Now(), Tool: "read_file"})
	}

	results := l.Search("read", 3)
	if len(results) != 3 {
		t.Errorf("got %d, want 3 (limited)", len(results))
	}
}

func TestHistory_All(t *testing.T) {
	tmp := t.TempDir()
	l, _ := NewLogger(tmp, "test", true, false)
	defer l.Close()

	for i := 0; i < 5; i++ {
		l.Log(Entry{Timestamp: time.Now(), Tool: "tool"})
	}

	history := l.History(0)
	if len(history) != 5 {
		t.Errorf("got %d, want 5 (0 = all)", len(history))
	}
}

func TestStats_Empty(t *testing.T) {
	tmp := t.TempDir()
	l, _ := NewLogger(tmp, "test", true, false)
	defer l.Close()

	stats := l.Stats()
	if len(stats) != 0 {
		t.Errorf("got %d stats, want 0", len(stats))
	}
}

func TestClose_DoubleClose(t *testing.T) {
	tmp := t.TempDir()
	l, _ := NewLogger(tmp, "test", true, false)

	l.Close()
	l.Close()
}
