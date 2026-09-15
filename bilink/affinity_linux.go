// Real per-core pinning for the native backend (cmd/multicore), on Linux --
// unlike macOS (see affinity_stub.go), Linux's sched_setaffinity actually
// does what its name says: it hard-pins the calling thread to one CPU, not
// just a scheduler hint. No cgo needed -- SYS_SCHED_SETAFFINITY is a plain
// syscall number the standard library already knows about.
//
//go:build linux && !js

package bilink

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

// cpuSetBytes matches glibc's default cpu_set_t size (1024 bits) -- more
// than enough for any BIL_CORE_HINT this backend will ever assign (it
// never exceeds runtime.NumCPU(), see ../cmd/multicore).
const cpuSetBytes = 128

// applyCoreAffinityHint reads BIL_CORE_HINT (a 0-based core index
// cmd/multicore assigns this node) and hard-pins this thread to it via
// sched_setaffinity(2). Must run after runtime.LockOSThread() (see
// bilink_native.go's init) -- pid 0 in this syscall means "the calling
// thread," which is only meaningful once this goroutine is locked to one.
func applyCoreAffinityHint() {
	hint := os.Getenv("BIL_CORE_HINT")
	if hint == "" {
		return
	}
	core, err := strconv.Atoi(hint)
	if err != nil || core < 0 || core >= cpuSetBytes*8 {
		fmt.Fprintf(os.Stderr, "bilink: ignoring invalid BIL_CORE_HINT %q: %v\n", hint, err)
		return
	}
	var mask [cpuSetBytes]byte
	mask[core/8] |= 1 << uint(core%8)
	_, _, errno := syscall.Syscall(syscall.SYS_SCHED_SETAFFINITY, 0, uintptr(cpuSetBytes), uintptr(unsafe.Pointer(&mask[0])))
	if errno != 0 {
		fmt.Fprintf(os.Stderr, "bilink: sched_setaffinity(core %d) failed: %v\n", core, errno)
	}
}
