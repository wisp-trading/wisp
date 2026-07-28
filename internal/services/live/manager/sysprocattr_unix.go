//go:build unix

package manager

import "syscall"

// detachSysProcAttr returns SysProcAttr so the strategy survives parent exit
// by running in its own process group.
func detachSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}
