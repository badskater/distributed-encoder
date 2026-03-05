package api

import (
	"fmt"
	"net/http"

	"github.com/badskater/distributed-encoder/internal/db"
)

// handleMetrics returns basic operational counters in Prometheus text format.
// This endpoint is unauthenticated so Prometheus scrapers can reach it without
// session credentials.
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	// --- Job counts by status ---
	fmt.Fprintln(w, "# HELP distencoder_jobs_total Number of jobs by status.")
	fmt.Fprintln(w, "# TYPE distencoder_jobs_total gauge")
	for _, st := range []string{"queued", "running", "completed", "failed", "cancelled"} {
		_, total, err := s.store.ListJobs(ctx, db.ListJobsFilter{Status: st, PageSize: 1})
		if err != nil {
			s.logger.Warn("metrics: list jobs", "status", st, "err", err)
			continue
		}
		fmt.Fprintf(w, "distencoder_jobs_total{status=%q} %d\n", st, total)
	}

	// --- Agent counts by status ---
	agents, err := s.store.ListAgents(ctx)
	if err != nil {
		s.logger.Warn("metrics: list agents", "err", err)
		return
	}

	counts := map[string]int{}
	for _, a := range agents {
		counts[a.Status]++
	}

	fmt.Fprintln(w, "# HELP distencoder_agents_total Number of registered agents by status.")
	fmt.Fprintln(w, "# TYPE distencoder_agents_total gauge")
	for _, st := range []string{"idle", "running", "offline", "draining", "pending_approval"} {
		fmt.Fprintf(w, "distencoder_agents_total{status=%q} %d\n", st, counts[st])
	}
}
