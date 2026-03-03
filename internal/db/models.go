package db

import "time"

// EncodeConfig holds the job-level configuration used to expand a queued job
// into individual tasks and to generate per-task script files.
type EncodeConfig struct {
	RunScriptTemplateID   string            `json:"run_script_template_id"`
	FrameserverTemplateID string            `json:"frameserver_template_id,omitempty"`
	ChunkBoundaries       []ChunkBoundary   `json:"chunk_boundaries"`
	OutputRoot            string            `json:"output_root"`
	OutputExtension       string            `json:"output_extension,omitempty"` // default "mkv"
	ExtraVars             map[string]string `json:"extra_vars,omitempty"`
}

// ChunkBoundary defines the inclusive frame range for one encoding task.
type ChunkBoundary struct {
	StartFrame int `json:"start_frame"`
	EndFrame   int `json:"end_frame"`
}

// The model types here mirror the database rows returned by queries.
// They are separate from the shared domain types (internal/shared) so the
// DB layer can be tested and evolved independently.

// User is a row from the users table.
type User struct {
	ID           string
	Username     string
	Email        string
	Role         string
	PasswordHash *string
	OIDCSub      *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Agent is a row from the agents table.
type Agent struct {
	ID            string
	Name          string
	Hostname      string
	IPAddress     string
	Status        string
	Tags          []string
	GPUVendor     string
	GPUModel      string
	GPUEnabled    bool
	AgentVersion  string
	OSVersion     string
	CPUCount      int32
	RAMMIB        int64
	NVENC         bool
	QSV           bool
	AMF           bool
	APIKeyHash    *string
	LastHeartbeat *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Source is a row from the sources table.
type Source struct {
	ID         string
	Filename   string
	UNCPath    string
	SizeBytes  int64
	DetectedBy *string
	State      string
	VMafScore  *float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Job is a row from the jobs table.
type Job struct {
	ID             string
	SourceID       string
	Status         string
	JobType        string
	Priority       int
	TargetTags     []string
	TasksTotal     int
	TasksPending   int
	TasksRunning   int
	TasksCompleted int
	TasksFailed    int
	EncodeConfig   EncodeConfig
	CompletedAt    *time.Time
	FailedAt       *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Task is a row from the tasks table.
type Task struct {
	ID            string
	JobID         string
	ChunkIndex    int
	Status        string
	AgentID       *string
	ScriptDir     string
	SourcePath    string
	OutputPath    string
	Variables     map[string]string
	ExitCode      *int
	FramesEncoded *int64
	AvgFPS        *float64
	OutputSize    *int64
	DurationSec   *int64
	VMafScore     *float64
	PSNR          *float64
	SSIM          *float64
	ErrorMsg      *string
	StartedAt     *time.Time
	CompletedAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// TaskLog is a row from the task_logs table.
type TaskLog struct {
	ID       int64
	TaskID   string
	JobID    string
	Stream   string
	Level    string
	Message  string
	Metadata map[string]any
	LoggedAt time.Time
}

// Template is a row from the templates table.
type Template struct {
	ID          string
	Name        string
	Description string
	Type        string
	Extension   string
	Content     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Variable is a row from the variables table.
type Variable struct {
	ID          string
	Name        string
	Value       string
	Description string
	Category    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Webhook is a row from the webhooks table.
type Webhook struct {
	ID        string
	Name      string
	Provider  string
	URL       string
	Secret    *string // raw HMAC-SHA256 signing key; nil means no signing
	Events    []string
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AnalysisResult is a row from the analysis_results table.
type AnalysisResult struct {
	ID        string
	SourceID  string
	Type      string
	FrameData []byte
	Summary   []byte
	CreatedAt time.Time
}

// Session is a row from the sessions table.
type Session struct {
	Token     string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// EnrollmentToken is a row from the enrollment_tokens table.
type EnrollmentToken struct {
	ID        string
	Token     string
	CreatedBy string
	UsedBy    *string
	UsedAt    *time.Time
	ExpiresAt time.Time
	CreatedAt time.Time
}

// ---------------------------------------------------------------------------
// Parameter structs — one per write operation
// ---------------------------------------------------------------------------

type CreateUserParams struct {
	Username     string
	Email        string
	Role         string
	PasswordHash *string
	OIDCSub      *string
}

type UpsertAgentParams struct {
	Name         string
	Hostname     string
	IPAddress    string
	Tags         []string
	GPUVendor    string
	GPUModel     string
	GPUEnabled   bool
	AgentVersion string
	OSVersion    string
	CPUCount     int32
	RAMMIB       int64
	NVENC        bool
	QSV          bool
	AMF          bool
}

type UpdateAgentHeartbeatParams struct {
	ID      string
	Status  string
	Metrics map[string]any
}

type CreateSourceParams struct {
	Filename   string
	UNCPath    string
	SizeBytes  int64
	DetectedBy *string
}

type ListSourcesFilter struct {
	State    string
	Cursor   string
	PageSize int
}

type CreateJobParams struct {
	SourceID     string
	JobType      string
	Priority     int
	TargetTags   []string
	EncodeConfig EncodeConfig
}

type ListJobsFilter struct {
	Status   string
	Cursor   string
	PageSize int
}

type CreateTaskParams struct {
	JobID      string
	ChunkIndex int
	SourcePath string
	OutputPath string
	Variables  map[string]string
}

type CompleteTaskParams struct {
	ID            string
	ExitCode      int
	FramesEncoded int64
	AvgFPS        float64
	OutputSize    int64
	DurationSec   int64
	VMafScore     *float64
	PSNR          *float64
	SSIM          *float64
}

type InsertTaskLogParams struct {
	TaskID   string
	JobID    string
	Stream   string
	Level    string
	Message  string
	Metadata map[string]any
	LoggedAt *time.Time
}

type ListTaskLogsParams struct {
	TaskID   string
	Stream   string
	Cursor   int64
	PageSize int
}

type CreateTemplateParams struct {
	Name        string
	Description string
	Type        string
	Extension   string
	Content     string
}

type UpdateTemplateParams struct {
	ID          string
	Name        string
	Description string
	Content     string
}

type UpsertVariableParams struct {
	Name        string
	Value       string
	Description string
	Category    string
}

type CreateWebhookParams struct {
	Name     string
	Provider string
	URL      string
	Secret   *string // raw HMAC-SHA256 signing key
	Events   []string
}

type UpdateWebhookParams struct {
	ID      string
	Name    string
	URL     string
	Events  []string
	Enabled bool
}

type InsertWebhookDeliveryParams struct {
	WebhookID    string
	Event        string
	Payload      []byte
	ResponseCode *int
	Success      bool
	Attempt      int
	ErrorMsg     *string
}

type UpsertAnalysisResultParams struct {
	SourceID  string
	Type      string
	FrameData []byte
	Summary   []byte
}

type CreateSessionParams struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
}

type ListJobLogsParams struct {
	JobID    string
	Stream   string
	Cursor   int64
	PageSize int
}

type CreateEnrollmentTokenParams struct {
	Token     string
	CreatedBy string
	ExpiresAt time.Time
}

type ConsumeEnrollmentTokenParams struct {
	Token   string
	AgentID string
}
