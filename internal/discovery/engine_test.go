// Package discovery_test contains unit tests for the discovery engine and
// all registered agent detectors.
package discovery_test

import (
	"testing"
	"time"

	"github.com/chethanyadav456/agentnetra/internal/discovery"
	"github.com/chethanyadav456/agentnetra/pkg/logger"
	"github.com/chethanyadav456/agentnetra/pkg/models"
)

// makeEngine returns a discovery Engine wired with a no-op logger for tests.
func makeEngine() *discovery.Engine {
	return discovery.NewEngine(logger.NewNop())
}

// proc builds a minimal ProcessInfo for use in table-driven detector tests.
func proc(pid int32, name, exe, cmdline string) models.ProcessInfo {
	return models.ProcessInfo{
		PID:            pid,
		Name:           name,
		ExecutablePath: exe,
		CommandLine:    cmdline,
	}
}

// ---------------------------------------------------------------------------
// DetectClaude
// ---------------------------------------------------------------------------

func TestDetectClaude(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		proc    models.ProcessInfo
		wantHit bool
	}{
		{"exact binary name", proc(1, "claude", "/usr/local/bin/claude", "claude"), true},
		{"claude-code binary", proc(2, "claude-code", "/usr/local/bin/claude-code", "claude-code"), true},
		{"versioned binary", proc(3, "claude-3.5", "/opt/claude/claude-3.5", "claude-3.5 --version"), true},
		{"exe path match", proc(4, "something", "/home/user/.local/bin/claude", "something"), true},
		{"cmdline prefix", proc(5, "sh", "/bin/sh", "claude --dangerously-skip-permissions"), true},
		{"no match", proc(6, "node", "/usr/bin/node", "node server.js"), false},
		{"cursor should not match", proc(7, "cursor", "/Applications/Cursor.app/cursor", "cursor ."), false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			eng := makeEngine()
			agents := eng.Discover([]models.ProcessInfo{tc.proc})
			got := len(agents) > 0 && agents[0].Name == "claude"

			if got != tc.wantHit {
				t.Errorf("DetectClaude(%q): got=%v, want=%v", tc.name, got, tc.wantHit)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DetectCursor
// ---------------------------------------------------------------------------

func TestDetectCursor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		proc    models.ProcessInfo
		wantHit bool
	}{
		{"exact name", proc(10, "cursor", "/Applications/Cursor.app/cursor", "cursor ."), true},
		{"cursor helper electron", proc(11, "cursor helper", "/Applications/Cursor.app/cursor helper", "cursor helper"), true},
		{"exe path", proc(12, "Electron", "/Applications/Cursor.app/Contents/MacOS/Electron", "Electron"), true},
		{"no match python", proc(13, "python3", "/usr/bin/python3", "python3 -m aider"), false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			eng := makeEngine()
			agents := eng.Discover([]models.ProcessInfo{tc.proc})
			got := len(agents) > 0 && agents[0].Name == "cursor"

			if got != tc.wantHit {
				t.Errorf("DetectCursor(%q): got=%v, want=%v", tc.name, got, tc.wantHit)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DetectGemini
// ---------------------------------------------------------------------------

func TestDetectGemini(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		proc    models.ProcessInfo
		wantHit bool
	}{
		{"exact binary", proc(20, "gemini", "/usr/local/bin/gemini", "gemini"), true},
		{"exe path", proc(21, "node", "/usr/local/lib/node_modules/gemini/bin/gemini", "node gemini"), true},
		{"cmdline prefix", proc(22, "sh", "/bin/sh", "gemini --version"), true},
		{"no match", proc(23, "aider", "/usr/local/bin/aider", "aider"), false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			eng := makeEngine()
			agents := eng.Discover([]models.ProcessInfo{tc.proc})
			got := len(agents) > 0 && agents[0].Name == "gemini"

			if got != tc.wantHit {
				t.Errorf("DetectGemini(%q): got=%v, want=%v", tc.name, got, tc.wantHit)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DetectAider
// ---------------------------------------------------------------------------

func TestDetectAider(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		proc    models.ProcessInfo
		wantHit bool
	}{
		{"exact binary", proc(30, "aider", "/home/user/.local/bin/aider", "aider"), true},
		{"python -m aider", proc(31, "python3", "/usr/bin/python3", "python3 -m aider"), true},
		{"python path aider", proc(32, "python", "/usr/bin/python", "python /home/user/.local/lib/aider/__main__.py"), true},
		{"no match", proc(33, "node", "/usr/bin/node", "node index.js"), false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			eng := makeEngine()
			agents := eng.Discover([]models.ProcessInfo{tc.proc})
			got := len(agents) > 0 && agents[0].Name == "aider"

			if got != tc.wantHit {
				t.Errorf("DetectAider(%q): got=%v, want=%v", tc.name, got, tc.wantHit)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DetectCodex
// ---------------------------------------------------------------------------

func TestDetectCodex(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		proc    models.ProcessInfo
		wantHit bool
	}{
		{"exact binary", proc(40, "codex", "/usr/local/bin/codex", "codex"), true},
		{"node invocation", proc(41, "node", "/usr/bin/node", "node /usr/local/lib/node_modules/codex/bin/codex"), true},
		{"exe path", proc(42, "node", "/usr/local/lib/node_modules/@openai/codex/bin/codex", "codex"), true},
		{"no match", proc(43, "python", "/usr/bin/python", "python server.py"), false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			eng := makeEngine()
			agents := eng.Discover([]models.ProcessInfo{tc.proc})
			got := len(agents) > 0 && agents[0].Name == "codex"

			if got != tc.wantHit {
				t.Errorf("DetectCodex(%q): got=%v, want=%v", tc.name, got, tc.wantHit)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DetectOpenHands
// ---------------------------------------------------------------------------

func TestDetectOpenHands(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		proc    models.ProcessInfo
		wantHit bool
	}{
		{"exact binary", proc(50, "openhands", "/usr/local/bin/openhands", "openhands"), true},
		{"python module", proc(51, "python3", "/usr/bin/python3", "python3 -m openhands.core.main"), true},
		{"docker cmdline", proc(52, "docker", "/usr/bin/docker", "docker run ghcr.io/all-hands-ai/openhands:0.14"), true},
		{"no match", proc(53, "cursor", "/Applications/Cursor.app/cursor", "cursor ."), false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			eng := makeEngine()
			agents := eng.Discover([]models.ProcessInfo{tc.proc})
			got := len(agents) > 0 && agents[0].Name == "openhands"

			if got != tc.wantHit {
				t.Errorf("DetectOpenHands(%q): got=%v, want=%v", tc.name, got, tc.wantHit)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Engine integration tests
// ---------------------------------------------------------------------------

func TestEngine_Discover_MultipleAgents(t *testing.T) {
	t.Parallel()

	eng := makeEngine()

	procs := []models.ProcessInfo{
		proc(100, "claude", "/usr/local/bin/claude", "claude"),
		proc(101, "aider", "/usr/local/bin/aider", "aider"),
		proc(102, "gemini", "/usr/local/bin/gemini", "gemini"),
		proc(103, "nginx", "/usr/sbin/nginx", "nginx -g daemon off;"),
	}

	agents := eng.Discover(procs)

	if len(agents) != 3 {
		t.Fatalf("Discover: expected 3 agents, got %d", len(agents))
	}

	found := map[string]bool{}
	for _, a := range agents {
		found[a.Name] = true
	}

	for _, name := range []string{"claude", "aider", "gemini"} {
		if !found[name] {
			t.Errorf("Discover: expected agent %q not found in results", name)
		}
	}
}

func TestEngine_Discover_SetsStatus(t *testing.T) {
	t.Parallel()

	eng := makeEngine()
	procs := []models.ProcessInfo{proc(200, "claude", "/usr/local/bin/claude", "claude")}

	agents := eng.Discover(procs)
	if len(agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(agents))
	}

	if agents[0].Status != "running" {
		t.Errorf("Status: got %q, want %q", agents[0].Status, "running")
	}
}

func TestEngine_Discover_SetsDiscoveredAt(t *testing.T) {
	t.Parallel()

	before := time.Now()
	eng := makeEngine()
	procs := []models.ProcessInfo{proc(300, "cursor", "/Applications/Cursor.app/cursor", "cursor")}

	agents := eng.Discover(procs)
	after := time.Now()

	if len(agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(agents))
	}

	if agents[0].DiscoveredAt.Before(before) || agents[0].DiscoveredAt.After(after) {
		t.Errorf("DiscoveredAt %v is outside expected range [%v, %v]",
			agents[0].DiscoveredAt, before, after)
	}
}

func TestEngine_Discover_EmptyProcessList(t *testing.T) {
	t.Parallel()

	eng := makeEngine()
	agents := eng.Discover(nil)

	if len(agents) != 0 {
		t.Errorf("expected 0 agents for empty input, got %d", len(agents))
	}
}

func TestEngine_Register_CustomDetector(t *testing.T) {
	t.Parallel()

	eng := makeEngine()
	eng.Register("myagent", func(name, exe, cmdline string) bool {
		return name == "myagent"
	})

	procs := []models.ProcessInfo{proc(400, "myagent", "/usr/bin/myagent", "myagent")}
	agents := eng.Discover(procs)

	if len(agents) != 1 || agents[0].Name != "myagent" {
		t.Errorf("custom detector: expected 1 match for 'myagent', got %v", agents)
	}
}
