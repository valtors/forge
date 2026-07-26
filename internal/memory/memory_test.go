package memory

import (
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	s, err := New("")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
}

func TestRemember(t *testing.T) {
	s, _ := New("")
	defer s.Close()

	if err := s.Remember("alice", "uses", "linux"); err != nil {
		t.Fatal(err)
	}
	if err := s.Remember("bob", "reports_to", "alice"); err != nil {
		t.Fatal(err)
	}
}

func TestRecall(t *testing.T) {
	s, _ := New("")
	defer s.Close()

	s.Remember("alice", "uses", "linux")
	s.Remember("bob", "reports_to", "alice")

	facts, err := s.Recall("alice", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(facts) != 2 {
		t.Errorf("got %d facts, want 2", len(facts))
	}
}

func TestRecall_NoMatch(t *testing.T) {
	s, _ := New("")
	defer s.Close()

	s.Remember("alice", "uses", "linux")

	facts, err := s.Recall("nonexistent", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(facts) != 0 {
		t.Errorf("got %d facts, want 0", len(facts))
	}
}

func TestForget(t *testing.T) {
	s, _ := New("")
	defer s.Close()

	s.Remember("alice", "uses", "linux")
	s.Forget("alice", "uses")

	facts, err := s.Recall("alice", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(facts) != 0 {
		t.Errorf("got %d facts after forget, want 0", len(facts))
	}
}

func TestAll(t *testing.T) {
	s, _ := New("")
	defer s.Close()

	s.Remember("alice", "uses", "linux")
	s.Remember("bob", "uses", "macos")

	facts, err := s.All()
	if err != nil {
		t.Fatal(err)
	}
	if len(facts) != 2 {
		t.Errorf("got %d facts, want 2", len(facts))
	}
}

func TestNew_FileDB(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.db")
	s, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if err := s.Remember("test", "has", "value"); err != nil {
		t.Fatal(err)
	}

	s2, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()

	facts, _ := s2.Recall("test", 10)
	if len(facts) != 1 {
		t.Errorf("got %d facts after reopen, want 1", len(facts))
	}
}

func TestNew_InvalidPath(t *testing.T) {
	_, err := New("/nonexistent/path/that/does/not/exist/db.db")
	if err == nil {
		t.Error("should error on invalid path")
	}
}

func TestRecall_LimitOne(t *testing.T) {
	s, _ := New("")
	defer s.Close()

	s.Remember("alice", "uses", "linux")
	s.Remember("alice", "likes", "go")

	facts, err := s.Recall("alice", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(facts) != 1 {
		t.Errorf("got %d, want 1 (limit)", len(facts))
	}
}

func TestRemember_Duplicate(t *testing.T) {
	s, _ := New("")
	defer s.Close()

	s.Remember("alice", "uses", "linux")
	s.Remember("alice", "uses", "linux")

	facts, _ := s.All()
	if len(facts) != 2 {
		t.Errorf("got %d facts, want 2 (duplicates allowed)", len(facts))
	}
}

func TestForget_NotFound(t *testing.T) {
	s, _ := New("")
	defer s.Close()

	err := s.Forget("nonexistent", "predicate")
	if err != nil {
		t.Errorf("forget nonexistent should not error: %v", err)
	}
}

func TestAll_Empty(t *testing.T) {
	s, _ := New("")
	defer s.Close()

	facts, err := s.All()
	if err != nil {
		t.Fatal(err)
	}
	if len(facts) != 0 {
		t.Errorf("got %d, want 0", len(facts))
	}
}

func TestAll_AfterForget(t *testing.T) {
	s, _ := New("")
	defer s.Close()

	s.Remember("alice", "uses", "linux")
	s.Remember("bob", "uses", "macos")
	s.Forget("alice", "uses")

	facts, _ := s.All()
	if len(facts) != 1 {
		t.Errorf("got %d, want 1 after forget", len(facts))
	}
	if facts[0].Subject != "bob" {
		t.Errorf("remaining fact subject = %s, want bob", facts[0].Subject)
	}
}
