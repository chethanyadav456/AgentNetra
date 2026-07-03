// Package process_test contains unit and integration tests for the process
// enumeration engine.
package process_test

import (
	"os"
	"testing"

	"github.com/chethanyadav456/agentnetra/internal/process"
	"github.com/chethanyadav456/agentnetra/pkg/logger"
)

// newTestEnumerator creates an Enumerator backed by a no-op logger for testing.
func newTestEnumerator(t *testing.T) *process.Enumerator {
	t.Helper()
	return process.NewEnumerator(logger.NewNop())
}

// NOTE: The process integration tests below do NOT use t.Parallel().
// gopsutil/v4 on macOS performs lazy initialisation of Darwin syscall
// function pointers via ebitengine/purego. Concurrent first-calls to
// process.Exe() race on that initialisation. This is an upstream bug
// (tracked in gopsutil issue #1552). Running these tests sequentially
// is the correct workaround until gopsutil bumps purego to ≥0.8.2.

func TestProcesses_ReturnsNonEmpty(t *testing.T) {
	enum := newTestEnumerator(t)

	procs, err := enum.Processes()
	if err != nil {
		t.Fatalf("Processes: unexpected error: %v", err)
	}

	if len(procs) == 0 {
		t.Error("Processes: expected at least one process, got 0")
	}
}

func TestProcesses_CurrentProcessIncluded(t *testing.T) {
	enum := newTestEnumerator(t)

	currentPID := int32(os.Getpid())

	procs, err := enum.Processes()
	if err != nil {
		t.Fatalf("Processes: unexpected error: %v", err)
	}

	for _, p := range procs {
		if p.PID == currentPID {
			// Verify mandatory fields are non-empty for our own process.
			if p.Name == "" {
				t.Error("own process: Name is empty")
			}
			return
		}
	}

	t.Errorf("Processes: current process (pid=%d) not found in result", currentPID)
}

func TestProcesses_FieldSanity(t *testing.T) {
	enum := newTestEnumerator(t)

	procs, err := enum.Processes()
	if err != nil {
		t.Fatalf("Processes: unexpected error: %v", err)
	}

	for _, p := range procs {
		if p.PID <= 0 {
			t.Errorf("process PID %d is non-positive", p.PID)
		}
		// Name must never be empty — buildProcessInfo returns an error otherwise.
		if p.Name == "" {
			t.Errorf("process PID %d has empty Name", p.PID)
		}
	}
}

func TestProcess_SinglePID(t *testing.T) {
	enum := newTestEnumerator(t)
	currentPID := int32(os.Getpid())

	info, err := enum.Process(currentPID)
	if err != nil {
		t.Fatalf("Process(%d): unexpected error: %v", currentPID, err)
	}

	if info.PID != currentPID {
		t.Errorf("Process: PID mismatch: got %d, want %d", info.PID, currentPID)
	}
	if info.Name == "" {
		t.Error("Process: Name is empty for current process")
	}
}

func TestProcess_InvalidPID_ReturnsError(t *testing.T) {
	t.Parallel() // This test never calls Exe(); it is safe to run in parallel.

	enum := newTestEnumerator(t)

	// PID 0 is not a valid user process PID.
	_, err := enum.Process(0)
	if err == nil {
		t.Error("Process(0): expected error for invalid PID, got nil")
	}
}
