//go:build unix

package manager_test

import (
	"syscall"

	. "github.com/onsi/gomega"
)

func expectDetachedSysProcAttr(attr *syscall.SysProcAttr) {
	Expect(attr.Setpgid).To(BeTrue())
	// 0 means create a new process group so the child survives parent exit.
	Expect(attr.Pgid).To(Equal(0))
}
