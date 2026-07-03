# AgentNetra

> **See What Your AI Agents Are Really Doing**

[![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8.svg)](https://go.dev)
[![Version](https://img.shields.io/badge/version-0.1.0-blue.svg)](https://github.com/chethanyadav456/agentnetra/releases)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)

AgentNetra is an open-source **observability, security, and governance platform** for autonomous AI agents. It gives developers complete visibility into which AI coding agents are running on their machine, how long they have been running, and what lifecycle events have occurred.

---

## Features (v0.1.0)

| Feature | Status |
|---------|--------|
| Agent Discovery | ✅ |
| Process Enumeration | ✅ |
| Claude Code Detection | ✅ |
| Cursor Detection | ✅ |
| Gemini CLI Detection | ✅ |
| Aider Detection | ✅ |
| Codex CLI Detection | ✅ |
| OpenHands Detection | ✅ |
| SQLite Storage | ✅ |
| Session Tracking | ✅ |
| Lifecycle Events | ✅ |
| CLI Interface | ✅ |

---

## Supported Platforms

- **Linux** (amd64, arm64)
- **macOS** (amd64, arm64 / Apple Silicon)
- **Windows** (amd64)

---

## Quick Start

### Prerequisites

- Go 1.24+
- No CGo required (uses `modernc.org/sqlite`)

### Install from source

```bash
git clone https://github.com/chethanyadav456/agentnetra.git
cd agentnetra
make install
```

### Usage

```bash
# Scan for AI agents currently running on your machine
agentnetra scan

# List detected AI agents (running only)
agentnetra agents

# List all agents including stopped ones
agentnetra agents --all

# List recorded sessions
agentnetra sessions

# List lifecycle events
agentnetra events

# Print version
agentnetra --version
```

### Example Output

```
$ agentnetra scan

AGENT    PID     PARENT PID   STATUS    DISCOVERED AT         COMMAND
-----    ---     ----------   ------    -------------         -------
claude   98321   97800        running   2026-07-03 16:10:05   claude --dangerously-skip-permissions
gemini   91245   91100        running   2026-07-03 16:09:40   gemini

Scan complete. Found 2 agent(s). Events: 2.
```

---

## Architecture

```
AgentNetra CLI
      │
      ▼
 Service Layer
      │
 ┌────┴────┬──────────┬──────────┐
 ▼         ▼          ▼          ▼
Process  Discovery  Storage   Events
Engine    Engine     Layer    Engine
      │
      ▼
Operating System
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for the full specification.

---

## Configuration

AgentNetra reads configuration from (in priority order):

1. Environment variables — prefix `AGENTNETRA_`
2. `~/.agentnetra/config.yaml`
3. Built-in defaults

Example environment override:

```bash
AGENTNETRA_LOG_LEVEL=debug agentnetra scan
```

---

## Development

```bash
# Run tests
make test

# Run tests with coverage report
make test-coverage

# Build binary
make build

# Format code
make fmt

# Vet code
make vet
```

---

## Database

AgentNetra stores data locally in SQLite:

| Platform | Default path |
|----------|-------------|
| Linux | `~/.agentnetra/agentnetra.db` |
| macOS | `~/Library/Application Support/AgentNetra/agentnetra.db` |
| Windows | `%APPDATA%/AgentNetra/agentnetra.db` |

No data is ever sent externally. All data remains on your machine.

---

## Roadmap

| Version | Feature |
|---------|---------|
| v0.1.0 | Agent Discovery (current) |
| v0.2.0 | Lifecycle Monitoring |
| v0.3.0 | Token Attribution |
| v0.4.0 | Process Tree Visualization |
| v0.5.0 | Security Engine |
| v0.6.0 | Policy Engine |
| v1.0.0 | Desktop Application |

---

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) before submitting a pull request.

---

## License

Apache License 2.0 — see [LICENSE](LICENSE) for details.
