# forge

[![CI](https://github.com/valtors/forge/actions/workflows/ci.yml/badge.svg)](https://github.com/valtors/forge/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-0F172A?style=flat-square)](./LICENSE)
[![Go](https://img.shields.io/badge/go-1.23+-00ADD8?style=flat-square)](https://go.dev/)
[![tests](https://img.shields.io/badge/tests-100-green?style=flat-square)]()

local-first agent runtime. one binary. memory, sandboxing, observation, security, package management. agents run on top.

## what

forge is the operating system for AI agents. you write a config file. forge handles everything else.

```
forge init my-agent
forge run my-agent.toml
forge list
forge status
forge validate my-agent.toml
forge remember my-agent alice uses linux
forge recall my-agent
forge forget my-agent alice uses
forge logs my-agent-1234567890
forge kill my-agent-1234567890
```

one binary. no cloud. no daemon. no build step. boring tech (Go, SQLite, TOML).

## why

today you wire 15 MCP servers into a config file by hand. you hope they work. you can't see what your agent does. you can't sandbox it. you can't give it memory that survives restarts. you can't manage multiple agents.

forge fixes all of that. it's the layer between you and the agent.

| | manual config | forge |
|---|---|---|
| start an agent | edit json, hope | `forge run agent.toml` |
| memory | none | cairn (sqlite, temporal) |
| sandboxing | none | vault (fs overlay, env sanitize, net policy) |
| observation | console.log | observer (trace tools, sqlite log) |
| MCP management | manual json | smith (install, compose, secure) |
| security | hope | mcprobe (prompt injection, tool shadowing) |
| multiple agents | impossible | `forge list`, `forge kill` |
| notifications | drowned | pulse (filtered, prioritized) |

forge ties the Valtors ecosystem together. each tool is a kernel module. forge is the kernel.

## config

```toml
[agent]
name = "my-agent"
model = "claude-sonnet"
command = "claude"

[memory]
backend = "cairn"
path = "~/.local/share/forge/my-agent/memory.db"

[sandbox]
enabled = true
allow = ["./src"]
net = ["github.com"]

[mcp]
servers = ["filesystem", "git", "valtors/cairn"]

[observe]
log = true
trace = true
```

## architecture

```
forge/
  internal/
    config/      TOML config, load, save, defaults
    agent/        process manager, start/stop/list agents
    sandbox/      fs overlay, env sanitizer, network policy
    observe/      tool call logger, trace history, search
    mcp/          MCP server resolver (npm, git, local)
    memory/       temporal sqlite store (remember, recall, forget)
  main.go         CLI entry point
  commands.go      command handlers (run, list, logs, kill, init)
```

/## memory commands

forge also manages agent memory directly from the CLI:

```
forge remember my-agent alice uses linux
forge recall my-agent
forge recall my-agent alice
forge forget my-agent alice uses
```

memory persists across restarts. stored in sqlite. one file per agent.

## what it does when you run `forge run`

1. loads your agent.toml
2. sets up sandbox (fs overlay, strips secrets from env, blocks paths)
3. initializes memory (sqlite, temporal store)
4. resolves and starts MCP servers (npm, git, local)
5. starts the agent process
6. wires observation (logs every tool call, stores in sqlite)
7. agent runs with all services wired up
8. you get logs, traces, memory, sandboxing -- all automatic

## install

```bash
go install github.com/valtors/forge@latest
```

or build from source:

```bash
git clone https://github.com/valtors/forge
cd forge
go build -o forge .
```

## usage

```bash
# create a new agent config
forge init my-agent

# edit the config
vim my-agent.toml

# run the agent with all services wired
forge run my-agent.toml

# see running agents
forge list

# view logs
forge logs my-agent-1234567890

# stop an agent
forge kill my-agent-1234567890
```

## what gets sandboxed

when sandbox is enabled:

- **filesystem**: agent gets an overlay dir. writes go there. reads only from allowlisted paths. `.ssh`, `.aws`, `.env`, `.gnupg` are invisible.
- **environment**: every env var matching token, secret, password, credential, api_key, auth, aws_, azure_, openai, anthropic, stripe, resend, database_url, private_key, github_token, gh_pat -- stripped. agent sees a clean shell.
- **network**: only allowlisted hosts. everything else refused. wildcard support (`*.npmjs.org`).

## the ecosystem

forge is the runtime. these are the modules:

- **cairn** -- agent memory. temporal knowledge store in sqlite. `cargo install cairn-memory`
- **vault** -- agent sandbox. fs overlay, env sanitizer, net policy. `go install github.com/valtors/vault`
- **observer** -- tool call logging. trace tools injected back to agent. `go install github.com/valtors/observer`
- **smith** -- MCP package manager. install, compose, secure. `cargo install smith-mcp`
- **mcprobe** -- MCP security scanner. prompt injection, tool shadowing, baseline drift. `go install github.com/tamish560/mcprobe`
- **pulse** -- notification agent. filtered, prioritized, remembered.
- **relay** -- one MCP server for files, web, PDFs, images. `npx userelay`
- **reflow** -- SSR-safe responsive toolkit for TypeScript. `npm install usereflow`

## tests

100 tests. 73.3% coverage. all pass.

```bash
go test ./... -count=1
```

## tech

go. single binary. sqlite (pure-go, no cgo for memory store). toml config. stdlib everything else. boring tech on purpose.

## license

MIT. strictly open source. no cloud tier, no enterprise plan, no proprietary fork.
