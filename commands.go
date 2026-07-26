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

func runCmd(args []string) int {
	cfgPath := ""
	if len(args) > 0 {
		cfgPath = args[0]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	dataDir := filepath.Join(config.DataDir(), "forge")
	logDir := filepath.Join(dataDir, "logs")
	sbx := sandbox.New(cfg.Sandbox.Enabled, cfg.Sandbox.Allow, cfg.Sandbox.Net)
	agentID := fmt.Sprintf("%s-%d", cfg.Agent.Name, time.Now().Unix())

	if err := sbx.Setup(); err != nil {
		fmt.Fprintf(os.Stderr, "sandbox setup: %v\n", err)
		return 1
	}
	defer sbx.Cleanup()

	var mem *memory.Store
	if cfg.Memory.Backend == "cairn" || cfg.Memory.Backend == "" {
		mem, err = memory.New(cfg.Memory.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "memory init: %v\n", err)
			return 1
		}
		defer mem.Close()
	}

	logger, err := observe.NewLogger(logDir, agentID, cfg.Observe.Log, cfg.Observe.Trace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger init: %v\n", err)
		return 1
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
		return 1
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
	return 0
}

func listCmd(args []string) int {
	dataDir := filepath.Join(config.DataDir(), "forge")
	mgr := agent.NewManager(dataDir)

	agents := mgr.List()
	if len(agents) == 0 {
		fmt.Println("no running agents")
		return 0
	}

	for _, a := range agents {
		fmt.Printf("%-20s %s  pid=%d  up=%s\n",
			a.Name, a.ID, a.PID(), a.Uptime().Round(time.Second))
	}
	return 0
}

func logsCmd(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: forge logs <agent-id>")
		return 1
	}

	dataDir := filepath.Join(config.DataDir(), "forge")
	logPath := filepath.Join(dataDir, "logs", args[0]+".log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read logs: %v\n", err)
		return 1
	}
	fmt.Print(string(data))
	return 0
}

func killCmd(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: forge kill <agent-id>")
		return 1
	}

	dataDir := filepath.Join(config.DataDir(), "forge")
	mgr := agent.NewManager(dataDir)

	if err := mgr.Stop(args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "kill: %v\n", err)
		return 1
	}
	fmt.Printf("agent %s stopped\n", args[0])
	return 0
}

func initCmd(args []string) int {
	name := "my-agent"
	if len(args) > 0 {
		name = args[0]
	}

	cfg := config.Default(name)
	path := name + ".toml"
	if err := cfg.Save(path); err != nil {
		fmt.Fprintf(os.Stderr, "save: %v\n", err)
		return 1
	}

	fmt.Printf("created %s\n", path)
	fmt.Println("edit it, then run: forge run " + path)
	return 0
}

func statusCmd(args []string) int {
	dataDir := filepath.Join(config.DataDir(), "forge")
	mgr := agent.NewManager(dataDir)
	agents := mgr.List()

	fmt.Println("forge status")
	fmt.Println()

	if len(agents) == 0 {
		fmt.Println("agents: none running")
	} else {
		fmt.Printf("agents: %d running\n", len(agents))
		for _, a := range agents {
			fmt.Printf("  %s  pid=%d  up=%s  status=%s\n",
				a.ID, a.PID(), a.Uptime().Round(time.Second), a.Status())
		}
	}

	logDir := filepath.Join(dataDir, "logs")
	if entries, err := os.ReadDir(logDir); err == nil {
		fmt.Printf("logs: %d files\n", len(entries))
	} else {
		fmt.Println("logs: none")
	}

	memDir := filepath.Join(dataDir)
	if entries, err := os.ReadDir(memDir); err == nil {
		dbCount := 0
		for _, e := range entries {
			if filepath.Ext(e.Name()) == ".db" {
				dbCount++
			}
		}
		fmt.Printf("memory stores: %d\n", dbCount)
	}

	return 0
}

func validateCmd(args []string) int {
	cfgPath := ""
	if len(args) > 0 {
		cfgPath = args[0]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid: %v\n", err)
		return 1
	}

	fmt.Println("config valid")
	fmt.Printf("  agent: %s (model: %s)\n", cfg.Agent.Name, cfg.Agent.Model)
	fmt.Printf("  sandbox: %v (%d paths, %d hosts)\n",
		cfg.Sandbox.Enabled, len(cfg.Sandbox.Allow), len(cfg.Sandbox.Net))
	fmt.Printf("  memory: %s\n", cfg.Memory.Backend)
	fmt.Printf("  mcp: %d servers\n", len(cfg.MCP.Servers))
	fmt.Printf("  observe: log=%v trace=%v\n", cfg.Observe.Log, cfg.Observe.Trace)

	fmt.Printf("  blocked paths: %d patterns\n", len(sandbox.BlockedPaths()))
	fmt.Printf("  secret patterns: %d\n", len(sandbox.SecretPatterns()))

	return 0
}

func rememberCmd(args []string) int {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: forge remember <agent-name> <subject> <predicate> <object>")
		return 1
	}

	agentName := args[0]
	subject := args[1]
	predicate := args[2]
	object := ""
	if len(args) > 3 {
		object = args[3]
	}

	dataDir := filepath.Join(config.DataDir(), "forge")
	memPath := filepath.Join(dataDir, agentName, "memory.db")

	mem, err := memory.New(memPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "memory: %v\n", err)
		return 1
	}
	defer mem.Close()

	if err := mem.Remember(subject, predicate, object); err != nil {
		fmt.Fprintf(os.Stderr, "remember: %v\n", err)
		return 1
	}

	fmt.Printf("remembered: %s %s %s\n", subject, predicate, object)
	return 0
}

func recallCmd(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: forge recall <agent-name> [query]")
		return 1
	}

	agentName := args[0]
	query := ""
	if len(args) > 1 {
		query = args[1]
	}

	dataDir := filepath.Join(config.DataDir(), "forge")
	memPath := filepath.Join(dataDir, agentName, "memory.db")

	mem, err := memory.New(memPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "memory: %v\n", err)
		return 1
	}
	defer mem.Close()

	if query == "" {
		facts, err := mem.All()
		if err != nil {
			fmt.Fprintf(os.Stderr, "recall: %v\n", err)
			return 1
		}
		for _, f := range facts {
			fmt.Printf("%s %s %s\n", f.Subject, f.Predicate, f.Object)
		}
		return 0
	}

	facts, err := mem.Recall(query, 20)
	if err != nil {
		fmt.Fprintf(os.Stderr, "recall: %v\n", err)
		return 1
	}
	for _, f := range facts {
		fmt.Printf("%s %s %s\n", f.Subject, f.Predicate, f.Object)
	}
	return 0
}

func forgetCmd(args []string) int {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: forge forget <agent-name> <subject> <predicate>")
		return 1
	}

	agentName := args[0]
	subject := args[1]
	predicate := args[2]

	dataDir := filepath.Join(config.DataDir(), "forge")
	memPath := filepath.Join(dataDir, agentName, "memory.db")

	mem, err := memory.New(memPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "memory: %v\n", err)
		return 1
	}
	defer mem.Close()

	if err := mem.Forget(subject, predicate); err != nil {
		fmt.Fprintf(os.Stderr, "forget: %v\n", err)
		return 1
	}

	fmt.Printf("forgotten: %s %s\n", subject, predicate)
	return 0
}
