// Package discovery — agent detector implementations.
//
// Each exported function in this file implements the Detector signature:
//
//	func(name, exe, cmdline string) bool
//
// All three arguments have already been lowercased and trimmed by the engine.
// Detectors should be as precise as possible to minimise false positives while
// remaining resilient to version-specific binary name changes.
package discovery

import "strings"

// DetectClaude returns true when the process is Claude Code (Anthropic).
//
// Signatures:
//   - Binary named "claude" or "claude-code"
//   - Executable path containing "claude" (handles versioned paths like
//     "claude-3.5")
//   - Command line containing "claude" as the first token
func DetectClaude(name, exe, cmdline string) bool {
	// Direct binary name matches.
	if name == "claude" || name == "claude-code" {
		return true
	}
	// Versioned executable (e.g., claude-3, claude-code-1.2.3).
	if strings.HasPrefix(name, "claude-") {
		return true
	}
	// Executable path contains the word "claude".
	if strings.Contains(exe, "claude") {
		return true
	}
	// Command line starts with the claude binary.
	if strings.HasPrefix(cmdline, "claude") {
		return true
	}
	return false
}

// DetectCursor returns true when the process is the Cursor AI editor.
//
// Signatures:
//   - Binary named "cursor"
//   - Executable path containing "cursor" (Electron app)
//   - Command line contains "cursor" as the primary token
func DetectCursor(name, exe, cmdline string) bool {
	if name == "cursor" {
		return true
	}
	if strings.Contains(exe, "cursor") {
		return true
	}
	// Electron Cursor processes on macOS appear as "Cursor Helper" etc.
	if strings.HasPrefix(name, "cursor") {
		return true
	}
	return false
}

// DetectGemini returns true when the process is the Gemini CLI (Google).
//
// Signatures:
//   - Binary named "gemini"
//   - Executable path containing "gemini"
//   - Command line starting with "gemini"
func DetectGemini(name, exe, cmdline string) bool {
	if name == "gemini" {
		return true
	}
	if strings.Contains(exe, "gemini") {
		return true
	}
	if strings.HasPrefix(cmdline, "gemini") {
		return true
	}
	return false
}

// DetectAider returns true when the process is Aider (paul-gauthier/aider).
//
// Signatures:
//   - Binary named "aider"
//   - Python invocation: "python ... aider" or "python3 ... aider"
//   - Command line contains "aider" token after python interpreter
func DetectAider(name, exe, cmdline string) bool {
	if name == "aider" {
		return true
	}
	if strings.Contains(exe, "aider") {
		return true
	}
	// Aider is often invoked as: python -m aider  OR  python /path/to/aider/__main__.py
	if (strings.HasPrefix(name, "python") || strings.HasPrefix(exe, "python")) &&
		strings.Contains(cmdline, "aider") {
		return true
	}
	return false
}

// DetectCodex returns true when the process is OpenAI Codex CLI.
//
// Signatures:
//   - Binary named "codex"
//   - Executable path containing "codex"
//   - Node.js invocation containing "codex" in the command line
func DetectCodex(name, exe, cmdline string) bool {
	if name == "codex" {
		return true
	}
	if strings.Contains(exe, "codex") {
		return true
	}
	// Codex CLI is a Node.js package; may appear as: node /path/to/codex/bin/codex
	if (name == "node" || strings.HasPrefix(exe, "node")) &&
		strings.Contains(cmdline, "codex") {
		return true
	}
	return false
}

// DetectOpenHands returns true when the process is an OpenHands agent
// (formerly OpenDevin).
//
// Signatures:
//   - Binary named "openhands"
//   - Python process that includes "openhands" in the command line
//   - Docker / container process with "openhands" in the command
func DetectOpenHands(name, exe, cmdline string) bool {
	if name == "openhands" {
		return true
	}
	if strings.Contains(exe, "openhands") {
		return true
	}
	// Python module invocation: python -m openhands.core.main
	if strings.HasPrefix(name, "python") && strings.Contains(cmdline, "openhands") {
		return true
	}
	// Docker-based: docker run ... ghcr.io/all-hands-ai/openhands
	if strings.Contains(cmdline, "openhands") {
		return true
	}
	return false
}
