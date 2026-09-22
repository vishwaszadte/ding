//go:build windows

package daemon

import (
	"os/exec"
	"syscall"
)

func detachProcess(cmd *exec.Cmd) {
	// CREATE_NEW_PROCESS_GROUP = 0x00000200
	// DETACHED_PROCESS = 0x00000008
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x00000008 | 0x00000200,
		HideWindow:    true,
	}
}
