package engine

import (
	"context"
	"fmt"
)

// checkStaleAgents marks agents as offline if they haven't sent a heartbeat
// within the configured threshold.
func (e *Engine) checkStaleAgents(ctx context.Context) error {
	n, err := e.store.MarkStaleAgents(ctx, e.cfg.StaleThreshold)
	if err != nil {
		return fmt.Errorf("engine: mark stale agents: %w", err)
	}
	if n > 0 {
		e.logger.Info("marked stale agents offline", "count", n)
	}
	return nil
}
