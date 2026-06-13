//go:build windows

package monitor

import (
	"os/exec"
	"syscall"
)

const detachedProcess = 0x00000008

// prepareCommand detaches the child from the parent's console so GUI apps
// do not block waiting for stdin when watchdog runs inside cmd.exe.
func prepareCommand(cmd *exec.Cmd) {
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | detachedProcess,
	}
}
