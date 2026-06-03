package backend

import (
	"context"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

const (
	defaultStopTimeout = 5 * time.Second
	maxRestartAttempts = 3
)

type ProcessManager struct {
	cmdName string
	cmdArgs []string
	env     []string
	stdout  io.Writer
	stderr  io.Writer

	mu           sync.Mutex
	cmd          *exec.Cmd
	running      bool
	restartCount int
}

func NewProcessManager(name string, args []string, env []string) *ProcessManager {
	return &ProcessManager{
		cmdName: name,
		cmdArgs: args,
		env:     env,
	}
}

func (pm *ProcessManager) Start(ctx context.Context) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	cmd := exec.CommandContext(ctx, pm.cmdName, pm.cmdArgs...)
	cmd.Env = pm.env
	if pm.stdout != nil {
		cmd.Stdout = pm.stdout
	}
	if pm.stderr != nil {
		cmd.Stderr = pm.stderr
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	pm.cmd = cmd
	pm.running = true
	return nil
}

func (pm *ProcessManager) Stop(ctx context.Context) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.cmd == nil || !pm.running {
		return nil
	}

	pm.cmd.Process.Signal(os.Interrupt)

	done := make(chan error, 1)
	go func() {
		done <- pm.cmd.Wait()
	}()

	select {
	case <-done:
	case <-ctx.Done():
		pm.cmd.Process.Kill()
		<-done
	}

	pm.running = false
	pm.cmd = nil
	return nil
}

func (pm *ProcessManager) Running() bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.running
}

func (pm *ProcessManager) Restart(ctx context.Context) error {
	if err := pm.Stop(ctx); err != nil {
		return err
	}
	return pm.Start(ctx)
}
