package engine

import (
	"context"
	"log/slog"
	"time"

	"github.com/badskater/distributed-encoder/internal/db"
)

// Config holds the settings for the background engine loop.
type Config struct {
	DispatchInterval time.Duration
	StaleThreshold   time.Duration
	ScriptBaseDir    string
}

// Engine orchestrates job expansion and stale-agent detection on a timer.
type Engine struct {
	store  db.Store
	gen    *ScriptGenerator
	cfg    Config
	logger *slog.Logger
}

// New creates an Engine. Does not start the background loop.
func New(store db.Store, cfg Config, logger *slog.Logger) *Engine {
	return &Engine{
		store:  store,
		gen:    newScriptGenerator(store, cfg.ScriptBaseDir, logger),
		cfg:    cfg,
		logger: logger,
	}
}

// Start launches the background dispatch loop in a goroutine.
// Returns immediately. The loop runs until ctx is cancelled.
func (e *Engine) Start(ctx context.Context) {
	go e.loop(ctx)
}

func (e *Engine) loop(ctx context.Context) {
	ticker := time.NewTicker(e.cfg.DispatchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := e.expandPendingJobs(ctx); err != nil {
				e.logger.Warn("engine: expand pending jobs", "error", err)
			}
			if err := e.checkStaleAgents(ctx); err != nil {
				e.logger.Warn("engine: check stale agents", "error", err)
			}
		}
	}
}
