//go:build windows

package monitor

import (
	"syscall"
)

const (
	stillActive                      = 259
	processQueryLimitedInformation   = 0x1000
)

func processAlive(pid int) bool {
	handle, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)

	var exitCode uint32
	if err := syscall.GetExitCodeProcess(handle, &exitCode); err != nil {
		return false
	}
	return exitCode == stillActive
}
