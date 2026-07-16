//go:build linux && amd64

package execution

import "syscall"

func memfdCreateSyscall() uintptr { return 319 }

var _ = syscall.SYS_FCNTL
