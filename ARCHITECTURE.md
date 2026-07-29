# Architecture

## Overview

Forge is a local-first agent runtime. One binary manages the full lifecycle of coding agents: process supervision, configuration, memory, sandboxing, MCP server wiring, and observability. No daemon. No cloud. No build step.

```
                    +-----------+
                    |  forge    |
                    |  CLI      |
                    +-----+-----+
                          |
          +---------------+---------------+
          |               |               |
     +----v----+   +------v------+   +----v----+
     | config  |   |   agent     |   |  memory  |
     | (TOML)  |   |  (process   |   | (SQLite) |
     |         |   |   manager)  |   |          |
     +---------+   +------+------+   +----------+
                         |
          +--------------+--------------+
          |              |              |
     +----v---+   +------v------+  +----v----+
     |  mcp   |   |  sandbox    |  | observe |
     | (proxy)|   | (fs + env)  |  | (logger)|
     +--------+   +-------------+  +---------+
```

## Design Principles

1. **One binary, zero dependencies.** Forge ships as a single Go binary with CGO for SQLite. No runtime install, no service registration, no container required.
2. **Local-first.** All state lives on disk under `$XDG_DATA_HOME/forge/`. Memory, logs, and traces never leave the machine.
3. **Composable, not monolithic.** Each subsystem (memory, sandbox, MCP, observe) is a separate package with its own interface. Swap the memory backend from SQLite to Cairn by changing one config line.
4. **Boring tech.** Go stdlib, SQLite, TOML. No frameworks, no ORMs, no message queues.

## Components

### config (`internal/config`)

TOML-based agent configuration. One file per agent.

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
servers = ["filesystem", "git"]

[observe]
log = true
trace = true
```

The `config.Load` function validates required fields and applies defaults via `config.Default`. Configs are plain files -- no database, no schema migration. To change an agent's behavior, edit the file and re-run.

### agent (`internal/agent`)

Process manager. Each agent runs as an `exec.Cmd` with stdout/stderr captured to a log file. The `Manager` tracks all running agents in a map protected by `sync.RWMutex`.

- `Start(id, name, command, args, logPath)` -- spawns the process, creates a log file, starts a goroutine to watch for unexpected exit (crash detection).
- `Stop(id)` -- sends `os.Interrupt`, waits 10 seconds, then sends `SIGKILL`.
- `List()` -- returns all agents with their status (starting, running, stopped, crashed).
- `Get(id)` / `GetPID(id)` -- lookup by ID or PID.

Crash detection: a goroutine calls `cmd.Wait()` and flips the status from `running` to `crashed` if the process exits without an explicit `Stop`. This distinguishes "I killed it" from "it died."

### mcp (`internal/mcp`)

MCP server resolver and launcher. Takes a server name and figures out how to run it:

| Input format | Resolution |
|---|---|
| `@scope/pkg` | `npx -y @scope/pkg` |
| `npm:pkg` | `npx -y pkg` |
| `git:org/repo` | `npx -y org/repo` |
| `github:org/repo` | `npx -y org/repo` |
| `/path/to/binary` | Direct execution |
| `filesystem` | `npx -y filesystem` (fallback) |

The resolved server runs as a child process with stdin/stdout/stderr disconnected. The `MCPConfig()` method produces the JSON config block that gets injected into the agent's config.

### memory (`internal/memory`)

SQLite-backed fact store. Triple store: (subject, predicate, object). Soft delete via `closed_at` timestamp -- facts are never physically deleted, just marked closed.

Schema:
```sql
CREATE TABLE facts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    subject TEXT NOT NULL,
    predicate TEXT NOT NULL,
    object TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    closed_at DATETIME
);
CREATE INDEX idx_facts_subject ON facts(subject);
CREATE INDEX idx_facts_predicate ON facts(predicate);
```

- `Remember(s, p, o)` -- insert a fact.
- `Recall(query, limit)` -- LIKE search across subject, predicate, object. Only returns open facts.
- `Forget(s, p)` -- sets `closed_at` on matching facts. No hard delete.
- `All()` -- returns all open facts, newest first.

The memory backend is swappable via config. The default is SQLite; `backend = "cairn"` delegates to Cairn's MCP server instead.

### sandbox (`internal/sandbox`)

Filesystem and environment isolation for agent processes.

**Path control:**
- Blocked paths: `.ssh`, `.aws`, `.gnupg`, `.docker`, `.kube`, `.config/gcloud`, `.config/gh`, `.npmrc`, `.pypirc`, `.netrc`, `.env`, `.gitconfig`. These are never accessible regardless of allow list.
- Allowed paths: explicit allowlist from config. Anything outside the allowlist is denied.
- Resolution: absolute path comparison. `AllowedPath` resolves to abs, checks blocked patterns first, then checks allowlist.

**Environment sanitization:**
- Secret patterns: `TOKEN`, `SECRET`, `PASSWORD`, `CREDENTIAL`, `API_KEY`, `AUTH`, `AWS_`, `AZURE_`, `OPENAI`, `ANTHROPIC`, `CLAUPE`, `STRIPE`, `RESEND`, `MAILGUN`, `SENDGRID`, `DATABASE_URL`, `DSN`, `PRIVATE_KEY`, `NPM_TOKEN`, `GITHUB_TOKEN`, `GH_PAT`, `GOOGLE_API`.
- `SanitizeEnv` strips any env var whose key contains one of these patterns (case-insensitive). The agent never sees the host's secrets.

**Network control:**
- `AllowedHost(host)` checks against the `net` allowlist from config. Supports exact match and `*.domain` wildcards.
- Empty net list = all hosts allowed (fail open for local dev).

### observe (`internal/observe`)

Structured logger for agent tool calls. Each entry records: timestamp, agent ID, tool name, input, output, duration, error.

- In-memory ring buffer (`entries []Entry`) for fast `History()` and `Search()` queries.
- File-backed append log for persistence (`agentID.log` under the log directory).
- `trace` flag controls whether input/output is included in log lines (default: false = metadata only).
- `Stats()` returns per-tool call counts.

Truncation: input and output are capped at 200 chars in log lines. The in-memory buffer retains full content for `History()` and `Search()`.

## Data Flow

```
1. User: forge run agent.toml
2. config.Load("agent.toml") -> *Config
3. sandbox.New(cfg.Sandbox).Setup() -> overlay dir
4. memory.New(cfg.Memory.Path) -> *Store (SQLite)
5. For each cfg.MCP.Servers:
      mcp.Resolve(name) -> *Server
      server.Start() -> *exec.Cmd
6. agent.Manager.Start(id, name, cmd, args, logPath) -> agent process
7. observe.NewLogger(logDir, agentID, log, trace) -> *Logger
8. Agent runs. Tool calls logged via observe.Logger.Log(entry).
9. Memory accessible via remember/recall/forget commands.
10. User: forge kill agent
    -> Manager.Stop(id) -> SIGINT -> wait 10s -> SIGKILL
    -> Logger.Close()
    -> sandbox.Cleanup()
```

## Process Model

Forge itself is a short-lived CLI process. It starts agents as child processes, writes their state to disk, and exits. A subsequent `forge list` reads the state directory to reconstruct the agent list. This means:

- No daemon to keep alive.
- No IPC socket to manage.
- State survives forge crashes (it is on disk).
- The agent process is the source of truth for liveness. `forge list` checks if the PID is still running.

## File Layout

```
$XDG_DATA_HOME/forge/
  agent-name/
    memory.db          # SQLite memory store
    agent.log          # Observe log
    state.json         # Agent metadata (ID, PID, start time, status)
```

## Testing

100 tests across all packages. Coverage: 73.3%. Untested paths are `main()` (blocks on os.Exit) and `runCmd()` (blocks on signal handler). Package-level coverage:

| Package | Coverage |
|---|---|
| mcp | 100% |
| agent | 94% |
| observe | 92.7% |
| sandbox | 91.2% |
| config | 84% |
| memory | 81.4% |

## Dependencies

- `github.com/BurntSushi/toml` -- TOML parsing
- `github.com/mattn/go-sqlite3` -- SQLite (CGO, bundled)
- Go stdlib for everything else

No web framework. No CLI framework. No logging library. No test framework beyond `testing` + `testify/assert`.
