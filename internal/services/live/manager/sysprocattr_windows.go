//go:build windows

package manager

import "syscall"

// detachSysProcAttr returns SysProcAttr so the strategy is not tied to the
// parent's console control group (best-effort Windows analogue of Setpgid).
func detachSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
