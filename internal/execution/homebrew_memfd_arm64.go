//go:build linux && arm64

package execution

func memfdCreateSyscall() uintptr { return 279 }
