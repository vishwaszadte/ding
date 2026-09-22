package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// PIDPath returns the file path where the daemon process ID is recorded (~/.ding/ding.pid).
func PIDPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ding", "ding.pid"), nil
}

// LogPath returns the file path where daemon output is logged (~/.ding/daemon.log).
func LogPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ding", "daemon.log"), nil
}

// IsRunning checks whether the daemon process is actively running.
func IsRunning() (bool, int) {
	pidFile, err := PIDPath()
	if err != nil {
		return false, 0
	}

	data, err := os.ReadFile(pidFile)
	if err != nil {
		return false, 0
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return false, 0
	}

	if !isProcessAlive(pid) {
		// Process is dead, clean up stale PID file
		_ = os.Remove(pidFile)
		return false, 0
	}

	return true, pid
}

// SavePID records the current process ID into ~/.ding/ding.pid.
func SavePID() error {
	pidFile, err := PIDPath()
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(pidFile), 0755)
	return os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0644)
}

// RemovePID deletes the PID file on clean daemon shutdown.
func RemovePID() {
	pidFile, err := PIDPath()
	if err == nil {
		_ = os.Remove(pidFile)
	}
}

// AutoStart ensures the background daemon process is running.
// If not running, it spawns `ding daemon run` detached in the background!
func AutoStart() error {
	running, _ := IsRunning()
	if running {
		return nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine executable path: %w", err)
	}

	logFile, err := LogPath()
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(logFile), 0755)

	out, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("could not open daemon log file: %w", err)
	}

	cmd := exec.Command(exePath, "daemon", "run")
	cmd.Stdout = out
	cmd.Stderr = out

	// Detach process so it survives when terminal closes
	detachProcess(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start background daemon: %w", err)
	}

	// Release process resources so parent doesn't wait
	_ = cmd.Process.Release()
	return nil
}

// Stop terminates the running daemon process.
func Stop() error {
	running, pid := IsRunning()
	if !running {
		return fmt.Errorf("daemon is not currently running")
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if err := process.Kill(); err != nil {
		return fmt.Errorf("failed to kill daemon process: %w", err)
	}

	RemovePID()
	return nil
}
