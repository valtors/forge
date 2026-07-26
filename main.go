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
	args := os.Args[2:]

	var code int
	switch cmd {
	case "run":
		code = runCmd(args)
	case "list":
		code = listCmd(args)
	case "logs":
		code = logsCmd(args)
	case "kill":
		code = killCmd(args)
	case "init":
		code = initCmd(args)
	case "status":
		code = statusCmd(args)
	case "validate":
		code = validateCmd(args)
	case "remember":
		code = rememberCmd(args)
	case "recall":
		code = recallCmd(args)
	case "forget":
		code = forgetCmd(args)
	case "version":
		fmt.Println("forge 0.1.0")
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		printHelp()
		code = 1
	}

	os.Exit(code)
}

func printHelp() {
	fmt.Print(`forge - local-first agent runtime

usage:
  forge init [name]        create an agent config
  forge run [config]       start an agent with all services wired
  forge list               show running agents
  forge status             show forge status (agents, logs, memory)
  forge validate [config]  validate an agent config
  forge remember <agent> <s> <p> <o>  store a fact
  forge recall <agent> [query]        query agent memory
  forge forget <agent> <s> <p>       remove a fact
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
