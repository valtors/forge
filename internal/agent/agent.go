package agent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type Status string

const (
	StatusRunning  Status = "running"
	StatusStopped  Status = "stopped"
	StatusCrashed  Status = "crashed"
	StatusStarting Status = "starting"
)

type Agent struct {
	ID        string
	Name      string
	Config    map[string]interface{}
	cmd       *exec.Cmd
	status    Status
	startedAt time.Time
	logFile   *os.File
	mu        sync.Mutex
}

type Manager struct {
	agents  map[string]*Agent
	mu      sync.RWMutex
	dataDir string
}

func NewManager(dataDir string) *Manager {
	return &Manager{
		agents:  make(map[string]*Agent),
		dataDir: dataDir,
	}
}

func (m *Manager) Start(id, name, command string, args []string, logPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.agents[id]; exists {
		return fmt.Errorf("agent %s already running", id)
	}

	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}

	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("create log file: %w", err)
	}

	cmd := exec.Command(command, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	a := &Agent{
		ID:        id,
		Name:      name,
		cmd:       cmd,
		status:    StatusStarting,
		startedAt: time.Now(),
		logFile:   logFile,
	}

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("start agent: %w", err)
	}

	a.status = StatusRunning
	m.agents[id] = a

	go func() {
		_ = cmd.Wait()
		a.mu.Lock()
		if a.status == StatusRunning {
			a.status = StatusCrashed
		}
		a.mu.Unlock()
	}()

	return nil
}

func (m *Manager) Stop(id string) error {
	m.mu.RLock()
	a, exists := m.agents[id]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agent %s not found", id)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cmd.Process != nil {
		_ = a.cmd.Process.Signal(os.Interrupt)
		done := make(chan error, 1)
		go func() { done <- a.cmd.Wait() }()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			_ = a.cmd.Process.Kill()
		}
	}

	a.status = StatusStopped
	if a.logFile != nil {
		a.logFile.Close()
	}
	return nil
}

func (m *Manager) List() []*Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agents := make([]*Agent, 0, len(m.agents))
	for _, a := range m.agents {
		agents = append(agents, a)
	}
	return agents
}

func (m *Manager) Get(id string) (*Agent, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.agents[id]
	return a, ok
}

func (a *Agent) Status() Status {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.status
}

func (a *Agent) Uptime() time.Duration {
	return time.Since(a.startedAt)
}

func (a *Agent) PID() int {
	if a.cmd.Process != nil {
		return a.cmd.Process.Pid
	}
	return 0
}

func (m *Manager) GetPID(id string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.agents[id]
	if !ok {
		return 0
	}
	return a.PID()
}
