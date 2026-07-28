//go:build windows

package manager

import "syscall"

// stillActive is the Windows exit code meaning the process has not exited.
const stillActive = 259

// processAlive reports whether a process with the given PID is still running.
// Wait4/Signal(0) are Unix-only; on Windows we OpenProcess + GetExitCodeProcess.
func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	// Match os.FindProcess access rights so liveness aligns with FindProcess/Kill.
	const da = syscall.STANDARD_RIGHTS_READ |
		syscall.PROCESS_QUERY_INFORMATION |
		syscall.SYNCHRONIZE
	h, err := syscall.OpenProcess(da, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(h)

	var code uint32
	if err := syscall.GetExitCodeProcess(h, &code); err != nil {
		return false
	}
	return code == stillActive
}
