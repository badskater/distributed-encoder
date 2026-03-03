package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	agentcfg "github.com/badskater/distributed-encoder/internal/agent/config"
	pb "github.com/badskater/distributed-encoder/internal/proto/encoderv1"
)

const (
	defaultConfigPath = `C:\ProgramData\distributed-encoder\agent.yaml`
	serviceName       = "distributed-encoder-agent"
)

// Run is the agent service entrypoint. It loads configuration, then either
// runs as a Windows Service or in the foreground depending on the execution
// context.
func Run(args []string) error {
	configPath := defaultConfigPath
	if len(args) > 0 && args[0] != "" {
		configPath = args[0]
	}

	cfg, err := agentcfg.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config from %s: %w", configPath, err)
	}

	log := slog.Default()

	runFn := func(ctx context.Context) error {
		r := &runner{
			cfg:   cfg,
			log:   log,
			state: pb.AgentState_AGENT_STATE_IDLE,
		}
		return r.run(ctx)
	}

	if isWindowsService() {
		log.Info("starting as Windows Service", "name", serviceName)
		return runAsWindowsService(serviceName, runFn)
	}

	// Foreground mode: handle interrupt signals.
	log.Info("starting in foreground mode")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	return runFn(ctx)
}
