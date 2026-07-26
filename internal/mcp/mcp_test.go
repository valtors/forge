package mcp

import (
	"os"
	"testing"
)

func TestResolve_NPM(t *testing.T) {
	srv, err := Resolve("@modelcontextprotocol/server-filesystem")
	if err != nil {
		t.Fatal(err)
	}
	if srv.Path != "npx" {
		t.Errorf("path = %s, want npx", srv.Path)
	}
	if len(srv.Args) < 2 || srv.Args[1] != "@modelcontextprotocol/server-filesystem" {
		t.Errorf("args = %v", srv.Args)
	}
}

func TestResolve_NPMPrefix(t *testing.T) {
	srv, err := Resolve("npm:@some/server")
	if err != nil {
		t.Fatal(err)
	}
	if srv.Args[1] != "@some/server" {
		t.Errorf("args = %v", srv.Args)
	}
}

func TestResolve_GitPrefix(t *testing.T) {
	srv, err := Resolve("git:valtors/cairn")
	if err != nil {
		t.Fatal(err)
	}
	if srv.Args[1] != "valtors/cairn" {
		t.Errorf("args = %v", srv.Args)
	}
}

func TestResolve_GitHubPrefix(t *testing.T) {
	srv, err := Resolve("github:valtors/observer")
	if err != nil {
		t.Fatal(err)
	}
	if srv.Args[1] != "valtors/observer" {
		t.Errorf("args = %v", srv.Args)
	}
}

func TestResolve_Default(t *testing.T) {
	srv, err := Resolve("some-random-server")
	if err != nil {
		t.Fatal(err)
	}
	if srv.Path != "npx" {
		t.Errorf("path = %s", srv.Path)
	}
}

func TestServer_String(t *testing.T) {
	srv, _ := Resolve("@modelcontextprotocol/filesystem")
	s := srv.String()
	if s == "" {
		t.Error("String should not be empty")
	}
}

func TestServer_MCPConfig(t *testing.T) {
	srv, _ := Resolve("@modelcontextprotocol/filesystem")
	cfg := srv.MCPConfig()
	if cfg["command"] != "npx" {
		t.Errorf("command = %v", cfg["command"])
	}
}

func TestServer_Start_Echo(t *testing.T) {
	srv, _ := Resolve("echo")
	srv.Path = "echo"
	srv.Args = []string{"hello"}

	cmd, err := srv.Start()
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Process == nil {
		t.Fatal("process should not be nil")
	}
	cmd.Process.Kill()
}

func TestServer_Start_InvalidCommand(t *testing.T) {
	srv := &Server{
		Name: "nonexistent-xyz",
		Path: "nonexistent-command-xyz-12345",
		Args: []string{},
	}

	_, err := srv.Start()
	if err == nil {
		t.Error("should error on invalid command")
	}
}

func TestServer_MCPConfig_Full(t *testing.T) {
	srv := &Server{
		Name: "test-server",
		Path: "/usr/local/bin/server",
		Args: []string{"--port", "8080"},
	}

	cfg := srv.MCPConfig()
	if cfg["command"] != "/usr/local/bin/server" {
		t.Errorf("command = %v", cfg["command"])
	}
	args, ok := cfg["args"].([]string)
	if !ok || len(args) != 2 {
		t.Errorf("args = %v", cfg["args"])
	}
}

func TestResolve_LocalPath(t *testing.T) {
	tmp := t.TempDir()
	serverPath := tmp + "/myserver"
	os.WriteFile(serverPath, []byte("#!/bin/sh\n"), 0755)

	srv, err := Resolve(serverPath)
	if err != nil {
		t.Fatal(err)
	}
	if srv.Path != serverPath && srv.Path != tmp+"/myserver" {
		t.Errorf("path = %s", srv.Path)
	}
}
