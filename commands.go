package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/valtors/forge/internal/agent"
	"github.com/valtors/forge/internal/config"
	mcpregistry "github.com/valtors/forge/internal/mcp"
	"github.com/valtors/forge/internal/memory"
	"github.com/valtors/forge/internal/observe"
	"github.com/valtors/forge/internal/sandbox"
)

func runCmd(args []string) {
	cfgPath := ""
	if len(args) > 0 {
		cfgPath = args[0]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	dataDir := filepath.Join(config.DataDir(), "forge")
	logDir := filepath.Join(dataDir, "logs")
	agentID := fmt.Sprintf("%s-%d", cfg.Agent.Name, time.Now().Unix())

	sbx := sandbox.New(cfg.Sandbox.Enabled, cfg.Sandbox.Allow, cfg.Sandbox.Net)
	if err := sbx.Setup(); err != nil {
		fmt.Fprintf(os.Stderr, "sandbox setup: %v\n", err)
		os.Exit(1)
	}
	defer sbx.Cleanup()

	var mem *memory.Store
	if cfg.Memory.Backend == "cairn" || cfg.Memory.Backend == "" {
		mem, err = memory.New(cfg.Memory.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "memory init: %v\n", err)
			os.Exit(1)
		}
		defer mem.Close()
	}

	logger, err := observe.NewLogger(logDir, agentID, cfg.Observe.Log, cfg.Observe.Trace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger init: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	for _, serverName := range cfg.MCP.Servers {
		srv, err := mcpregistry.Resolve(serverName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "resolve MCP server %s: %v\n", serverName, err)
			continue
		}
		logger.Log(observe.Entry{
			Timestamp: time.Now(),
			AgentID:   agentID,
			Tool:      "mcp.resolve",
			Input:     serverName,
			Output:    srv.String(),
		})
	}

	mgr := agent.NewManager(dataDir)

	command := cfg.Agent.Command
	if command == "" {
		command = "claude"
	}

	fmt.Printf("starting agent %s (id: %s)\n", cfg.Agent.Name, agentID)
	fmt.Printf("  model: %s\n", cfg.Agent.Model)
	fmt.Printf("  sandbox: %v\n", cfg.Sandbox.Enabled)
	fmt.Printf("  memory: %s\n", cfg.Memory.Backend)
	fmt.Printf("  mcp servers: %d\n", len(cfg.MCP.Servers))
	fmt.Printf("  observe: log=%v trace=%v\n", cfg.Observe.Log, cfg.Observe.Trace)
	fmt.Println()

	if err := mgr.Start(agentID, cfg.Agent.Name, command, []string{}, filepath.Join(logDir, agentID+".stdout")); err != nil {
		fmt.Fprintf(os.Stderr, "start: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("agent running. pid: %d\n", mgr.GetPID(agentID))
	fmt.Println("press Ctrl+C to stop")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	fmt.Println("\nstopping agent...")
	if err := mgr.Stop(agentID); err != nil {
		fmt.Fprintf(os.Stderr, "stop: %v\n", err)
	}
	fmt.Println("stopped.")
}

func listCmd(args []string) {
	dataDir := filepath.Join(config.DataDir(), "forge")
	mgr := agent.NewManager(dataDir)

	agents := mgr.List()
	if len(agents) == 0 {
		fmt.Println("no running agents")
		return
	}

	for _, a := range agents {
		fmt.Printf("%-20s %s  pid=%d  up=%s\n",
			a.Name, a.ID, a.PID(), a.Uptime().Round(time.Second))
	}
}

func logsCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: forge logs <agent-id>")
		os.Exit(1)
	}

	dataDir := filepath.Join(config.DataDir(), "forge")
	logPath := filepath.Join(dataDir, "logs", args[0]+".log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read logs: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(string(data))
}

func killCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: forge kill <agent-id>")
		os.Exit(1)
	}

	dataDir := filepath.Join(config.DataDir(), "forge")
	mgr := agent.NewManager(dataDir)

	if err := mgr.Stop(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "kill: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("agent %s stopped\n", args[0])
}

func initCmd(args []string) {
	name := "my-agent"
	if len(args) > 0 {
		name = args[0]
	}

	cfg := config.Default(name)
	path := name + ".toml"
	if err := cfg.Save(path); err != nil {
		fmt.Fprintf(os.Stderr, "save: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("created %s\n", path)
	fmt.Println("edit it, then run: forge run " + path)
}
