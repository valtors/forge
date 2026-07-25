package observe

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Entry struct {
	Timestamp time.Time
	AgentID   string
	Tool      string
	Input     string
	Output    string
	Duration  time.Duration
	Error     string
}

type Logger struct {
	mu      sync.Mutex
	entries []Entry
	file    *os.File
	enabled bool
	trace   bool
}

func NewLogger(logDir, agentID string, enabled, trace bool) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}

	path := filepath.Join(logDir, agentID+".log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	return &Logger{
		file:    f,
		enabled: enabled,
		trace:   trace,
	}, nil
}

func (l *Logger) Log(entry Entry) {
	if !l.enabled {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.entries = append(l.entries, entry)

	line := fmt.Sprintf("[%s] %s tool=%s dur=%s",
		entry.Timestamp.Format(time.RFC3339),
		entry.AgentID,
		entry.Tool,
		entry.Duration,
	)
	if entry.Error != "" {
		line += " error=" + entry.Error
	}
	if l.trace {
		if entry.Input != "" {
			line += " in=" + truncate(entry.Input, 200)
		}
		if entry.Output != "" {
			line += " out=" + truncate(entry.Output, 200)
		}
	}
	line += "\n"

	if l.file != nil {
		_, _ = l.file.WriteString(line)
	}
}

func (l *Logger) History(limit int) []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()

	if limit <= 0 || limit > len(l.entries) {
		limit = len(l.entries)
	}

	start := len(l.entries) - limit
	if start < 0 {
		start = 0
	}

	result := make([]Entry, limit)
	copy(result, l.entries[start:])
	return result
}

func (l *Logger) Search(query string, limit int) []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()

	query = strings.ToLower(query)
	var results []Entry
	for _, e := range l.entries {
		if strings.Contains(strings.ToLower(e.Tool), query) ||
			strings.Contains(strings.ToLower(e.Input), query) ||
			strings.Contains(strings.ToLower(e.Output), query) {
			results = append(results, e)
			if limit > 0 && len(results) >= limit {
				break
			}
		}
	}
	return results
}

func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func (l *Logger) Stats() map[string]int {
	l.mu.Lock()
	defer l.mu.Unlock()

	stats := make(map[string]int)
	for _, e := range l.entries {
		stats[e.Tool]++
	}
	return stats
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
