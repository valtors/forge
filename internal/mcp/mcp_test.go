package mcp

import (
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
