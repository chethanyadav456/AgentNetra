# Contributing to AgentNetra

Thank you for your interest in contributing to AgentNetra!

---

## Getting Started

### Prerequisites

- Go 1.24+
- Git

### Setup

```bash
git clone https://github.com/chethanyadav456/agentnetra.git
cd agentnetra
make tidy
make build
make test
```

---

## Development Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Make your changes
4. Run tests: `make test`
5. Run vet: `make vet`
6. Commit: `git commit -m "feat: describe your change"`
7. Push: `git push origin feat/my-feature`
8. Open a Pull Request

---

## Code Standards

- Follow idiomatic Go style
- All exported functions must have godoc comments
- New packages must have a package-level comment
- Unit tests required for all new functionality
- Target 80%+ test coverage
- No global state
- Use dependency injection

---

## Adding a New Agent Detector

1. Add the agent type constant to `pkg/constants/constants.go`
2. Add a `DetectXxx` function to `internal/discovery/detectors.go`
3. Register it in `internal/discovery/engine.go` `registerBuiltins()`
4. Add table-driven tests in `internal/discovery/engine_test.go`

---

## Commit Message Format

```
type: short description

types: feat | fix | docs | test | refactor | chore
```

---

## Reporting Bugs

Please open a GitHub issue with:

- OS and version
- Go version
- Steps to reproduce
- Expected vs actual behavior
