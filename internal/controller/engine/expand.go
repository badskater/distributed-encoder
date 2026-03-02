package engine

import (
	"context"
	"fmt"

	"github.com/badskater/distributed-encoder/internal/db"
)

// expandPendingJobs queries jobs needing expansion and expands each one.
func (e *Engine) expandPendingJobs(ctx context.Context) error {
	jobs, err := e.store.GetJobsNeedingExpansion(ctx)
	if err != nil {
		return fmt.Errorf("engine: get jobs needing expansion: %w", err)
	}
	for _, job := range jobs {
		if err := e.expandJob(ctx, job); err != nil {
			e.logger.Warn("engine: expand job failed",
				"job_id", job.ID,
				"error", err,
			)
		}
	}
	return nil
}

// expandJob creates tasks for a single job based on its chunk boundaries.
func (e *Engine) expandJob(ctx context.Context, job *db.Job) error {
	if len(job.EncodeConfig.ChunkBoundaries) == 0 {
		e.logger.Error("engine: job has no chunk boundaries, skipping",
			"job_id", job.ID,
		)
		return nil
	}

	source, err := e.store.GetSourceByID(ctx, job.SourceID)
	if err != nil {
		return fmt.Errorf("engine: get source %s: %w", job.SourceID, err)
	}

	ext := job.EncodeConfig.OutputExtension
	if ext == "" {
		ext = "mkv"
	}

	// Create a task for each chunk and render its scripts.
	for i := range job.EncodeConfig.ChunkBoundaries {
		// Build output path using string concatenation to preserve UNC prefix.
		outputPath := job.EncodeConfig.OutputRoot + `\` + job.ID + fmt.Sprintf(`\chunk_%04d.%s`, i, ext)

		task, err := e.store.CreateTask(ctx, db.CreateTaskParams{
			JobID:      job.ID,
			ChunkIndex: i,
			SourcePath: source.UNCPath,
			OutputPath: outputPath,
			Variables:  map[string]string{},
		})
		if err != nil {
			_ = e.store.UpdateJobStatus(ctx, job.ID, "failed")
			return fmt.Errorf("engine: create task chunk %d: %w", i, err)
		}

		scriptDir, err := e.gen.Render(ctx, job, task)
		if err != nil {
			_ = e.store.UpdateJobStatus(ctx, job.ID, "failed")
			return fmt.Errorf("engine: render scripts chunk %d: %w", i, err)
		}

		if err := e.store.SetTaskScriptDir(ctx, task.ID, scriptDir); err != nil {
			_ = e.store.UpdateJobStatus(ctx, job.ID, "failed")
			return fmt.Errorf("engine: set script dir chunk %d: %w", i, err)
		}
	}

	// All tasks created successfully — update counters and keep status as queued.
	if err := e.store.UpdateJobTaskCounts(ctx, job.ID); err != nil {
		return fmt.Errorf("engine: update task counts for job %s: %w", job.ID, err)
	}
	if err := e.store.UpdateJobStatus(ctx, job.ID, "queued"); err != nil {
		return fmt.Errorf("engine: update job status for job %s: %w", job.ID, err)
	}
	return nil
}
