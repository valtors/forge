package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	cmd := os.Args[1]
	switch cmd {
	case "run":
		runCmd(os.Args[2:])
	case "list":
		listCmd(os.Args[2:])
	case "logs":
		logsCmd(os.Args[2:])
	case "kill":
		killCmd(os.Args[2:])
	case "init":
		initCmd(os.Args[2:])
	case "version":
		fmt.Println("forge 0.1.0")
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`forge - local-first agent runtime

usage:
  forge init [name]        create an agent config
  forge run [config]       start an agent with all services wired
  forge list               show running agents
  forge logs [agent]       show logs for an agent
  forge kill [agent]       stop a running agent
  forge version            show version
  forge help               show this help

config (agent.toml):
  [agent]
  name = "my-agent"
  model = "claude-sonnet"

  [memory]
  backend = "cairn"

  [sandbox]
  enabled = true
  allow = ["./src"]
  net = ["github.com"]

  [mcp]
  servers = ["filesystem", "git"]

  [observe]
  log = true
  trace = true
`)
}
