//go:build unix

package manager

import (
	"os"
	"syscall"
)

// processAlive reports whether a process with the given PID is still running.
// On Unix, os.FindProcess always succeeds for any PID; Signal(0) is the real check.
// Wait4(WNOHANG) reaps a zombie child so a just-killed process is not reported alive.
func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return false
	}

	// If this PID is our unreaped child (zombie), Signal(0) can still succeed.
	// Reap it so we report death correctly after Kill without Cmd.Wait.
	var status syscall.WaitStatus
	wpid, err := syscall.Wait4(pid, &status, syscall.WNOHANG, nil)
	if err == nil && wpid == pid {
		return false
	}
	return true
}
