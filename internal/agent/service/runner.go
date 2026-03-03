package service

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	agentcfg "github.com/badskater/distributed-encoder/internal/agent/config"
	pb "github.com/badskater/distributed-encoder/internal/proto/encoderv1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const agentVersion = "0.1.0"

// runner holds the runtime state of the agent service.
type runner struct {
	cfg     *agentcfg.Config
	conn    *grpc.ClientConn
	client  pb.AgentServiceClient
	agentID string
	log     *slog.Logger
	offline *offlineStore

	mu            sync.Mutex
	state         pb.AgentState
	currentTaskID string
}

// run is the main lifecycle of the agent. It blocks until ctx is cancelled.
func (r *runner) run(ctx context.Context) error {
	r.log.Info("agent starting")

	// Open offline journal.
	offDB, err := newOfflineStore(r.cfg.Agent.OfflineDB)
	if err != nil {
		return fmt.Errorf("offline store: %w", err)
	}
	r.offline = offDB
	defer offDB.close()

	// Establish gRPC connection with reconnect loop.
	if err := r.connect(ctx); err != nil {
		return fmt.Errorf("initial connect: %w", err)
	}
	defer r.conn.Close()

	// Register with controller.
	if err := r.register(ctx); err != nil {
		return fmt.Errorf("register: %w", err)
	}

	// Sync any offline results from previous runs.
	r.syncOfflineResults(ctx)

	r.setState(pb.AgentState_AGENT_STATE_IDLE, "")

	// Start heartbeat goroutine.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		r.heartbeatLoop(ctx)
	}()

	// Task poll loop (runs in current goroutine).
	r.pollLoop(ctx)

	wg.Wait()
	r.log.Info("agent stopped")
	return nil
}

// connect establishes a gRPC connection to the controller using exponential
// backoff.
func (r *runner) connect(ctx context.Context) error {
	delay := r.cfg.Controller.Reconnect.InitialDelay
	if delay == 0 {
		delay = 5 * time.Second
	}
	maxDelay := r.cfg.Controller.Reconnect.MaxDelay
	if maxDelay == 0 {
		maxDelay = 5 * time.Minute
	}
	multiplier := r.cfg.Controller.Reconnect.Multiplier
	if multiplier < 1 {
		multiplier = 2.0
	}

	var creds grpc.DialOption
	if r.cfg.Controller.TLS.Cert != "" && r.cfg.Controller.TLS.Key != "" && r.cfg.Controller.TLS.CA != "" {
		tlsCfg, err := buildTLSConfig(r.cfg.Controller.TLS)
		if err != nil {
			return fmt.Errorf("tls config: %w", err)
		}
		creds = grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg))
	} else {
		r.log.Warn("TLS not configured, using insecure connection (dev mode)")
		creds = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	currentDelay := delay
	for {
		r.log.Info("connecting to controller", "address", r.cfg.Controller.Address)
		conn, err := grpc.NewClient(r.cfg.Controller.Address, creds)
		if err == nil {
			r.conn = conn
			r.client = pb.NewAgentServiceClient(conn)
			r.log.Info("connected to controller")
			return nil
		}
		r.log.Error("connection failed, retrying", "error", err, "delay", currentDelay)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(currentDelay):
		}

		currentDelay = time.Duration(float64(currentDelay) * multiplier)
		if currentDelay > maxDelay {
			currentDelay = maxDelay
		}
	}
}

// buildTLSConfig creates a mutual TLS configuration from the agent's cert,
// key, and CA files.
func buildTLSConfig(cfg agentcfg.TLSConfig) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(cfg.Cert, cfg.Key)
	if err != nil {
		return nil, fmt.Errorf("loading client cert/key: %w", err)
	}
	caCert, err := os.ReadFile(cfg.CA)
	if err != nil {
		return nil, fmt.Errorf("reading CA cert: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// register calls the controller Register RPC with exponential backoff until
// it succeeds or the context is cancelled.
func (r *runner) register(ctx context.Context) error {
	hostname := r.cfg.Agent.Hostname
	if hostname == "" {
		hostname, _ = os.Hostname()
	}
	ip := localIP()

	info := &pb.AgentInfo{
		Hostname:     hostname,
		IpAddress:    ip,
		AgentVersion: agentVersion,
		OsVersion:    runtime.GOOS + "/" + runtime.GOARCH,
		CpuCount:     int32(runtime.NumCPU()),
	}
	if r.cfg.GPU.Enabled {
		info.Gpu = &pb.GPUCapabilities{
			Vendor: r.cfg.GPU.Vendor,
			VramMb: int32(r.cfg.GPU.MaxVRAMMB),
		}
	}

	delay := r.cfg.Controller.Reconnect.InitialDelay
	if delay == 0 {
		delay = 5 * time.Second
	}
	maxDelay := r.cfg.Controller.Reconnect.MaxDelay
	if maxDelay == 0 {
		maxDelay = 5 * time.Minute
	}
	multiplier := r.cfg.Controller.Reconnect.Multiplier
	if multiplier < 1 {
		multiplier = 2.0
	}

	currentDelay := delay
	for {
		resp, err := r.client.Register(ctx, info)
		if err != nil {
			r.log.Error("registration failed, retrying", "error", err, "delay", currentDelay)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(currentDelay):
			}
			currentDelay = time.Duration(float64(currentDelay) * multiplier)
			if currentDelay > maxDelay {
				currentDelay = maxDelay
			}
			continue
		}
		if !resp.GetOk() {
			r.log.Warn("registration rejected", "message", resp.GetMessage())
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(currentDelay):
			}
			continue
		}
		r.agentID = resp.GetAgentId()
		r.log.Info("registered with controller", "agent_id", r.agentID, "approved", resp.GetApproved())
		return nil
	}
}

// heartbeatLoop sends periodic heartbeats to the controller.
func (r *runner) heartbeatLoop(ctx context.Context) {
	interval := r.cfg.Agent.HeartbeatInterval
	if interval == 0 {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.sendHeartbeat(ctx)
		}
	}
}

// sendHeartbeat sends a single heartbeat RPC.
func (r *runner) sendHeartbeat(ctx context.Context) {
	r.mu.Lock()
	state := r.state
	taskID := r.currentTaskID
	r.mu.Unlock()

	resp, err := r.client.Heartbeat(ctx, &pb.HeartbeatReq{
		AgentId:       r.agentID,
		State:         state,
		CurrentTaskId: taskID,
	})
	if err != nil {
		r.log.Error("heartbeat failed", "error", err)
		return
	}
	if resp.GetDrain() {
		r.log.Warn("controller requested drain")
	}
	if resp.GetDisabled() {
		r.log.Warn("controller disabled this agent")
	}
}

// pollLoop polls the controller for task assignments.
func (r *runner) pollLoop(ctx context.Context) {
	interval := r.cfg.Agent.PollInterval
	if interval == 0 {
		interval = 10 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.mu.Lock()
			busy := r.state == pb.AgentState_AGENT_STATE_BUSY
			r.mu.Unlock()
			if busy {
				continue
			}
			r.pollAndExecute(ctx)
		}
	}
}

// pollAndExecute polls for a task and, if one is assigned, executes it.
func (r *runner) pollAndExecute(ctx context.Context) {
	resp, err := r.client.PollTask(ctx, &pb.PollTaskReq{
		AgentId: r.agentID,
	})
	if err != nil {
		r.log.Error("poll task failed", "error", err)
		return
	}
	if !resp.GetHasTask() {
		return
	}

	r.log.Info("task received", "task_id", resp.GetTaskId(), "job_id", resp.GetJobId())
	r.setState(pb.AgentState_AGENT_STATE_BUSY, resp.GetTaskId())

	startedAt := time.Now()
	exitCode, execErr := r.executeTask(ctx, resp)
	completedAt := time.Now()

	success := execErr == nil && exitCode == 0
	errMsg := ""
	if execErr != nil {
		errMsg = execErr.Error()
	}

	result := &pb.TaskResult{
		TaskId:      resp.GetTaskId(),
		JobId:       resp.GetJobId(),
		Success:     success,
		ExitCode:    int32(exitCode),
		ErrorMsg:    errMsg,
		StartedAt:   timestamppb.New(startedAt),
		CompletedAt: timestamppb.New(completedAt),
	}

	if _, err := r.client.ReportResult(ctx, result); err != nil {
		r.log.Error("report result failed, saving offline", "error", err)
		if saveErr := r.offline.saveResult(resp.GetTaskId(), resp.GetJobId(), success, int32(exitCode), errMsg); saveErr != nil {
			r.log.Error("failed to save offline result", "error", saveErr)
		}
	} else {
		r.log.Info("task result reported", "task_id", resp.GetTaskId(), "success", success)
	}

	r.setState(pb.AgentState_AGENT_STATE_IDLE, "")
}

// executeTask writes script files to disk and runs the .bat entrypoint,
// streaming stdout/stderr to the controller.
func (r *runner) executeTask(ctx context.Context, task *pb.TaskAssignment) (int, error) {
	workDir := filepath.Join(r.cfg.Agent.WorkDir, task.GetTaskId())
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return -1, fmt.Errorf("creating work dir: %w", err)
	}

	// Write script files.
	var batPath string
	for name, content := range task.GetScripts() {
		p := filepath.Join(workDir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return -1, fmt.Errorf("creating script dir for %s: %w", name, err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			return -1, fmt.Errorf("writing script %s: %w", name, err)
		}
		if strings.HasSuffix(strings.ToLower(name), ".bat") {
			batPath = p
		}
	}
	if batPath == "" {
		return -1, fmt.Errorf("no .bat script found in task scripts")
	}

	// Apply task timeout if specified.
	execCtx := ctx
	if task.GetTimeoutSec() > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, time.Duration(task.GetTimeoutSec())*time.Second)
		defer cancel()
	}

	r.log.Info("executing task", "bat", batPath)
	cmd := exec.CommandContext(execCtx, "cmd.exe", "/c", batPath)
	cmd.Dir = workDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return -1, fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return -1, fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return -1, fmt.Errorf("starting command: %w", err)
	}

	// Open a log stream to the controller.
	logStream, streamErr := r.client.StreamLogs(ctx)

	// Stream stdout and stderr in goroutines.
	var streamWg sync.WaitGroup
	streamLine := func(stream string, scanner *bufio.Scanner) {
		defer streamWg.Done()
		for scanner.Scan() {
			line := scanner.Text()
			if logStream != nil && streamErr == nil {
				_ = logStream.Send(&pb.LogEntry{
					TaskId:    task.GetTaskId(),
					JobId:     task.GetJobId(),
					Stream:    stream,
					Level:     "info",
					Message:   line,
					Timestamp: timestamppb.Now(),
				})
			}
		}
	}

	streamWg.Add(2)
	go streamLine("stdout", bufio.NewScanner(stdout))
	go streamLine("stderr", bufio.NewScanner(stderr))

	streamWg.Wait()

	// Close the log stream.
	if logStream != nil && streamErr == nil {
		_, _ = logStream.CloseAndRecv()
	}

	waitErr := cmd.Wait()
	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return -1, waitErr
	}
	return 0, nil
}

// syncOfflineResults replays any buffered results to the controller.
func (r *runner) syncOfflineResults(ctx context.Context) {
	results, err := r.offline.pendingResults()
	if err != nil {
		r.log.Error("reading offline results", "error", err)
		return
	}
	if len(results) == 0 {
		return
	}

	r.log.Info("syncing offline results", "count", len(results))

	stream, err := r.client.SyncOfflineResults(ctx)
	if err != nil {
		r.log.Error("opening sync stream", "error", err)
		return
	}

	for _, res := range results {
		if err := stream.Send(&pb.TaskResult{
			TaskId:        res.TaskID,
			JobId:         res.JobID,
			Success:       res.Success,
			ExitCode:      res.ExitCode,
			ErrorMsg:      res.ErrorMsg,
			OfflineResult: true,
		}); err != nil {
			r.log.Error("sending offline result", "error", err, "task_id", res.TaskID)
			break
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		r.log.Error("closing sync stream", "error", err)
		return
	}

	// Mark synced results. Build a set of rejected IDs for quick lookup.
	rejected := make(map[string]bool, len(resp.GetRejectedTaskIds()))
	for _, tid := range resp.GetRejectedTaskIds() {
		rejected[tid] = true
	}
	for _, res := range results {
		// Mark both accepted and rejected as synced so we don't re-send.
		if err := r.offline.markSynced(res.ID); err != nil {
			r.log.Error("marking result synced", "error", err, "id", res.ID)
		}
	}

	r.log.Info("offline sync complete", "accepted", resp.GetAccepted(), "rejected", len(rejected))
}

// setState updates the agent's current state atomically.
func (r *runner) setState(state pb.AgentState, taskID string) {
	r.mu.Lock()
	r.state = state
	r.currentTaskID = taskID
	r.mu.Unlock()
}

// localIP returns the first non-loopback IPv4 address found on the host.
func localIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String()
		}
	}
	return "unknown"
}
