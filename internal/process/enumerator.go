// Package process provides cross-platform process enumeration and metadata
// collection using gopsutil. It is the lowest-level component of AgentNetra
// and has no knowledge of AI agents — it simply reflects the current OS
// process table as a slice of ProcessInfo value objects.
package process

import (
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
	"go.uber.org/zap"

	"github.com/chethanyadav456/agentnetra/pkg/logger"
	"github.com/chethanyadav456/agentnetra/pkg/models"
)

// Enumerator retrieves process information from the operating system.
// Use NewEnumerator to construct a valid instance.
type Enumerator struct {
	log *logger.Logger
}

// NewEnumerator constructs an Enumerator with the provided logger.
func NewEnumerator(log *logger.Logger) *Enumerator {
	return &Enumerator{log: log}
}

// Processes returns metadata for every process currently visible to the
// current user. Processes that cannot be inspected (e.g., due to permission
// errors) are silently skipped so that the enumeration is always as complete
// as possible.
func (e *Enumerator) Processes() ([]models.ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("process: enumerate processes: %w", err)
	}

	results := make([]models.ProcessInfo, 0, len(procs))

	for _, p := range procs {
		info, err := e.buildProcessInfo(p)
		if err != nil {
			// Permission denied and similar transient errors are expected on
			// most platforms for system processes. Log at debug and continue.
			e.log.Debug("process: skipping inaccessible process",
				zap.Int32("pid", p.Pid),
				zap.Error(err),
			)
			continue
		}
		results = append(results, info)
	}

	e.log.Debug("process: enumeration complete",
		zap.Int("total", len(results)),
	)

	return results, nil
}

// Process returns the ProcessInfo for a single PID. It returns an error if
// the process does not exist or cannot be inspected.
func (e *Enumerator) Process(pid int32) (models.ProcessInfo, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return models.ProcessInfo{}, fmt.Errorf("process: open process %d: %w", pid, err)
	}

	return e.buildProcessInfo(p)
}

// buildProcessInfo extracts all available metadata from a gopsutil Process
// handle and maps it to a models.ProcessInfo. Fields that are unavailable
// on a given platform are left at their zero value rather than causing the
// entire call to fail.
func (e *Enumerator) buildProcessInfo(p *process.Process) (models.ProcessInfo, error) {
	name, err := p.Name()
	if err != nil {
		return models.ProcessInfo{}, fmt.Errorf("process: read name for pid %d: %w", p.Pid, err)
	}

	ppid, err := p.Ppid()
	if err != nil {
		// Parent PID may be unavailable for some kernel threads.
		ppid = 0
	}

	exe, err := p.Exe()
	if err != nil {
		// Executable path can be unavailable for short-lived or system processes.
		exe = ""
	}

	cmdline, err := p.Cmdline()
	if err != nil {
		// Fall back to process name when command line is unavailable.
		cmdline = name
	}

	username, err := p.Username()
	if err != nil {
		username = ""
	}

	return models.ProcessInfo{
		PID:            p.Pid,
		ParentPID:      ppid,
		Name:           strings.TrimSpace(name),
		ExecutablePath: strings.TrimSpace(exe),
		CommandLine:    strings.TrimSpace(cmdline),
		Username:       strings.TrimSpace(username),
	}, nil
}
