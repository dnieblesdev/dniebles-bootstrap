//go:build linux && !amd64 && !arm64

package execution

// An invalid syscall number makes acquisition fail closed on unsupported Linux architectures.
func memfdCreateSyscall() uintptr { return ^uintptr(0) }
