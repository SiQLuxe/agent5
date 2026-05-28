package service

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type DebuggerService struct {
	process *exec.Cmd
	output  string
}

func NewDebuggerService() *DebuggerService {
	return &DebuggerService{}
}

func (ds *DebuggerService) RunCommand(cmd string, args []string, cwd string) (string, error) {
	command := exec.Command(cmd, args...)
	command.Dir = cwd

	var output strings.Builder
	var stderr strings.Builder

	command.Stdout = &output
	command.Stderr = &stderr

	err := command.Run()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}

	return output.String(), nil
}

func (ds *DebuggerService) StartProcess(cmd string, args []string, cwd string) error {
	if ds.process != nil && ds.process.Process != nil {
		ds.process.Process.Kill()
	}

	ds.process = exec.Command(cmd, args...)
	ds.process.Dir = cwd

	stdout, err := ds.process.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := ds.process.StderrPipe()
	if err != nil {
		return err
	}

	if err := ds.process.Start(); err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			ds.output += scanner.Text() + "\n"
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			ds.output += "ERROR: " + scanner.Text() + "\n"
		}
	}()

	return nil
}

func (ds *DebuggerService) StopProcess() error {
	if ds.process != nil && ds.process.Process != nil {
		return ds.process.Process.Kill()
	}
	return nil
}

func (ds *DebuggerService) GetOutput() string {
	return ds.output
}

func (ds *DebuggerService) ClearOutput() {
	ds.output = ""
}

func (ds *DebuggerService) IsRunning() bool {
	return ds.process != nil && ds.process.Process != nil && ds.process.ProcessState == nil
}

func (ds *DebuggerService) ExecuteGoCode(code string) (string, error) {
	tmpFile, err := os.CreateTemp("", "*.go")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(code); err != nil {
		return "", err
	}
	tmpFile.Close()

	return ds.RunCommand("go", []string{"run", tmpFile.Name()}, "")
}