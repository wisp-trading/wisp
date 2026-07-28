//go:build windows

package manager_test

import (
	"syscall"

	. "github.com/onsi/gomega"
)

func expectDetachedSysProcAttr(attr *syscall.SysProcAttr) {
	Expect(attr.CreationFlags & syscall.CREATE_NEW_PROCESS_GROUP).NotTo(BeZero())
}
