// No-op fallback for every native target that isn't Linux (see
// affinity_linux.go). macOS's own closest primitive,
// thread_policy_set(THREAD_AFFINITY_POLICY), was tried and dropped: it's
// only ever an advisory grouping hint even in principle, and confirmed
// directly to return KERN_NOT_SUPPORTED on every call on Apple Silicon --
// a pure no-op there, not worth the cgo dependency it would add. On
// macOS (and anywhere else without a real pin), multicore's proc-per-core
// mapping still holds in the sense that matters: one genuine OS process
// per placed proc, left to the OS scheduler to spread across cores like
// any other CPU-bound multi-process workload -- see ../../README.md's
// "Native core emulator" section.
//
//go:build !js && !linux

package bilink

func applyCoreAffinityHint() {}
