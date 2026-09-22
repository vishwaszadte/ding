//go:build !windows

package daemon

import (
	"os"
	"syscall"
)

// isProcessAlive checks whether a process is running on Unix using signal 0.
func isProcessAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}
