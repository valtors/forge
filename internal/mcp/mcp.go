package mcp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Server struct {
	Name string
	Path string
	Args []string
}

func Resolve(name string) (*Server, error) {
	if strings.HasPrefix(name, "@") || strings.HasPrefix(name, "npm:") {
		pkg := strings.TrimPrefix(name, "npm:")
		return &Server{
			Name: pkg,
			Path: "npx",
			Args: []string{"-y", pkg},
		}, nil
	}

	if strings.HasPrefix(name, "git:") || strings.HasPrefix(name, "github:") {
		pkg := strings.TrimPrefix(strings.TrimPrefix(name, "git:"), "github:")
		return &Server{
			Name: pkg,
			Path: "npx",
			Args: []string{"-y", pkg},
		}, nil
	}

	if _, err := os.Stat(name); err == nil {
		abs, _ := filepath.Abs(name)
		return &Server{
			Name: abs,
			Path: abs,
			Args: []string{},
		}, nil
	}

	return &Server{
		Name: name,
		Path: "npx",
		Args: []string{"-y", name},
	}, nil
}

func (s *Server) Start() (*exec.Cmd, error) {
	cmd := exec.Command(s.Path, s.Args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start MCP server %s: %w", s.Name, err)
	}
	return cmd, nil
}

func (s *Server) MCPConfig() map[string]interface{} {
	env := map[string]string{}
	return map[string]interface{}{
		"command": s.Path,
		"args":    s.Args,
		"env":     env,
	}
}

func (s *Server) String() string {
	return fmt.Sprintf("%s -> %s %s", s.Name, s.Path, strings.Join(s.Args, " "))
}
