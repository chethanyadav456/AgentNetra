// AgentNetra — CLI entry point.
//
// This file wires together configuration, logging, storage, and services,
// then hands control to the Cobra command tree. All subsystem initialisation
// lives here; commands contain only presentation logic.
package main

import (
	"fmt"
	"os"

	"github.com/chethanyadav456/agentnetra/cmd/agentnetra/commands"
	"github.com/chethanyadav456/agentnetra/internal/config"
	"github.com/chethanyadav456/agentnetra/internal/discovery"
	"github.com/chethanyadav456/agentnetra/internal/process"
	"github.com/chethanyadav456/agentnetra/internal/services"
	"github.com/chethanyadav456/agentnetra/internal/storage"
	"github.com/chethanyadav456/agentnetra/pkg/constants"
	"github.com/chethanyadav456/agentnetra/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "agentnetra: fatal error: %v\n", err)
		os.Exit(constants.ExitCodeError)
	}
}

// run is the real entry point, extracted for testability. It returns a
// non-nil error when startup fails; successful execution returns nil.
func run() error {
	// --- Configuration ---
	// The config file path is resolved by the root command flag parser before
	// any sub-command runs. For startup initialisation we load defaults first;
	// the root command will reload with the user-supplied path if provided.
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// --- Logger ---
	log, err := logger.New(logger.Config{
		Level:       cfg.Log.Level,
		Development: cfg.Log.Development,
		LogFile:     cfg.Log.File,
		AppName:     constants.AppName,
		AppVersion:  constants.AppVersion,
	})
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer func() { _ = log.Sync() }()

	// --- Storage ---
	db, err := storage.Open(cfg.Database.Path, log)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// --- Repositories ---
	agentRepo := storage.NewAgentRepository(db)
	sessionRepo := storage.NewSessionRepository(db)
	eventRepo := storage.NewEventRepository(db)

	// --- Engines ---
	enumerator := process.NewEnumerator(log)
	discEngine := discovery.NewEngine(log)

	// --- Services ---
	scanSvc := services.NewScanService(log, enumerator, discEngine, agentRepo, sessionRepo, eventRepo)
	agentListSvc := services.NewAgentListService(log, agentRepo)
	sessionListSvc := services.NewSessionListService(sessionRepo)
	eventListSvc := services.NewEventListService(eventRepo)

	// --- CLI ---
	deps := commands.Deps{
		Log:            log,
		ScanService:    scanSvc,
		AgentList:      agentListSvc,
		SessionList:    sessionListSvc,
		EventList:      eventListSvc,
	}

	root := commands.NewRootCmd(deps)
	return root.Execute()
}
