package service

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// offlineResult represents a task result buffered locally when the controller
// is unreachable.
type offlineResult struct {
	ID        int64
	TaskID    string
	JobID     string
	Success   bool
	ExitCode  int32
	ErrorMsg  string
	CreatedAt time.Time
}

// offlineStore wraps a SQLite database used to journal task results when the
// gRPC connection to the controller is unavailable.
type offlineStore struct {
	db *sql.DB
}

// newOfflineStore opens (or creates) the SQLite database at path and ensures
// the schema exists.
func newOfflineStore(path string) (*offlineStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening offline db: %w", err)
	}

	const schema = `CREATE TABLE IF NOT EXISTS offline_results (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id    TEXT    NOT NULL,
		job_id     TEXT    NOT NULL,
		success    INTEGER NOT NULL,
		exit_code  INTEGER NOT NULL,
		error_msg  TEXT    NOT NULL DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		synced     INTEGER NOT NULL DEFAULT 0
	)`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("creating offline_results table: %w", err)
	}

	return &offlineStore{db: db}, nil
}

// saveResult inserts a task result into the offline journal.
func (o *offlineStore) saveResult(taskID, jobID string, success bool, exitCode int32, errorMsg string) error {
	const q = `INSERT INTO offline_results (task_id, job_id, success, exit_code, error_msg) VALUES (?, ?, ?, ?, ?)`
	successInt := 0
	if success {
		successInt = 1
	}
	_, err := o.db.Exec(q, taskID, jobID, successInt, exitCode, errorMsg)
	return err
}

// pendingResults returns all results that have not yet been synced to the
// controller.
func (o *offlineStore) pendingResults() ([]offlineResult, error) {
	const q = `SELECT id, task_id, job_id, success, exit_code, error_msg, created_at FROM offline_results WHERE synced = 0 ORDER BY id`
	rows, err := o.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []offlineResult
	for rows.Next() {
		var r offlineResult
		var successInt int
		if err := rows.Scan(&r.ID, &r.TaskID, &r.JobID, &successInt, &r.ExitCode, &r.ErrorMsg, &r.CreatedAt); err != nil {
			return nil, err
		}
		r.Success = successInt != 0
		results = append(results, r)
	}
	return results, rows.Err()
}

// markSynced flags a result as successfully delivered to the controller.
func (o *offlineStore) markSynced(id int64) error {
	_, err := o.db.Exec(`UPDATE offline_results SET synced = 1 WHERE id = ?`, id)
	return err
}

// close releases the database connection.
func (o *offlineStore) close() error {
	return o.db.Close()
}
