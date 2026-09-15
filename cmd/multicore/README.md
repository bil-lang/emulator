# cmd/multicore — the native backend

Instead of one browser Web Worker per grid node, `multicore` runs one native OS process per grid node — the same one-node-is-one-real-execution-unit model [`cmd/wasm`](../wasm) uses, just on this machine's own CPU cores instead of in a browser tab. It's a small/native complement to the browser backend's larger grids, not a replacement: a launch is refused if `rows*cols` exceeds `runtime.NumCPU()`.

See the top-level [README.md](../../README.md) for how this fits alongside `cmd/wasm`, and for the shared list of node programs.

This works with zero changes to any node program — see [`bilink`](../../bilink)'s own README section: whichever platform you build a node program's `main.go`/`roles/*/main.go` for, Go's own build-tag resolution picks the matching `bilink` implementation automatically. `bilink_native.go`'s link primitive keeps the exact same genuine, unbuffered, two-sided rendezvous `cmd/wasm`'s `SharedArrayBuffer`+`Atomics` handshake provides, just over a pair of Unix-domain-socket connections per wired `link[i]` (one per direction, the direct analog of `index.html`'s `makeLinkPair()` ab/ba SharedArrayBuffer pair) instead of shared memory.

## Quickstart

Same on-disk shape `cmd/wasm`'s own Quickstart produces — build the roles, then run `multicore` instead of `cmd/wasm`, from the `emulator` repo root:

```
mkdir -p nodeprog/placedcontroller
../bil/tools/bilc/bilc ../bil/examples/20-placed-controller.bil nodeprog/placedcontroller/main.go
go run ./cmd/multicore -dir nodeprog/placedcontroller -rows 1 -cols 4
```

Or skip both steps and run `bil emu -target multicore examples/20-placed-controller.bil` from the sibling `bil` repo, which builds and launches this for you (and, unlike this binary's own `-rows`/`-cols`, can infer a grid size — see "Quirks" below).

`multicore` builds each role natively itself (no `GOOS`/`GOARCH` override) — there's no separate build step to run first, unlike `cmd/wasm`'s `GOOS=js GOARCH=wasm go build` line. A plain `main.go` with no placement (no `roles/deploy.json`) runs the same way, one process per grid position, all running the identical binary.

## Core mapping: real on Linux, scheduler-only on macOS

Each spawned process gets a `BIL_CORE_HINT` (a 0-based core index `multicore` assigns it), and `bilink_native.go` acts on it in `applyCoreAffinityHint()` (called after `runtime.LockOSThread()`, so the pin lands on the actual thread the node program runs on):

- **On Linux**, `affinity_linux.go` hard-pins the process to that exact core via `sched_setaffinity(2)` — a real pin, not a hint, and no cgo needed (`SYS_SCHED_SETAFFINITY` is a plain syscall number the standard library already knows).
- **On macOS**, there's no such call at all (`affinity_stub.go` is a no-op) — not because it wasn't tried: Darwin's closest primitive, `thread_policy_set(THREAD_AFFINITY_POLICY)`, was implemented first, and confirmed directly (via cgo, on an 8-core M-series Mac) to return `KERN_NOT_SUPPORTED` on every single call. It's not merely advisory there, it's inert, so it was dropped rather than kept as dead weight (and as a bonus, `bilink` now needs no cgo at all, on any platform).

In practice, on macOS this backend's real claim isn't "pins proc N to core N" — it's "spawns exactly one genuine OS process per placed proc, each independently schedulable, so a machine with enough cores actually runs them in parallel," and lets the OS scheduler do the spreading, same as any other CPU-bound multi-process workload. `nodeprog/corebench` (below) is how to see that this actually happens either way — with a real pin on Linux, or scheduler-driven spread on macOS.

## Seeing it happen: `corebench`

`nodeprog/corebench` is a LINPACK-style dense-matrix-multiply benchmark built to make this visible. Each column runs its own share of the kernel against a fixed 5-second wall-clock budget (not a fixed amount of work), so total run time stays roughly constant across any core count while throughput — block-multiplies completed — is what actually changes with it. The same compiled roles run at any width with no rebuild, so comparing core counts is just:

```
mkdir -p nodeprog/corebench
../bil/tools/bilc/bilc ../bil/examples/23-corebench.bil nodeprog/corebench/main.go
go run ./cmd/multicore -dir nodeprog/corebench -rows 1 -cols 1   # baseline
go run ./cmd/multicore -dir nodeprog/corebench -rows 1 -cols 4
go run ./cmd/multicore -dir nodeprog/corebench -rows 1 -cols 8
```

Watch Activity Monitor's (or `htop`'s) CPU history while each is running, and compare each run's final printed total. Measured directly on an 8-core M-series Mac: 1 column ≈ 4.6k block-multiplies/s, 4 columns ≈ 17.4k/s combined (≈3.8x), 8 columns ≈ 24.3k/s combined (≈5.3x) — scaling that flattens out past 4, consistent with that chip's 4-performance/4-efficiency core split rather than 8 identical cores, and a real example of what this backend can and can't tell you: it proves genuine concurrent core usage, not that every core is equally fast.

## Quirks

- **Grid size is capped at `runtime.NumCPU()`**, checked before anything spawns — this backend is for small, deliberately-sized grids, not the browser backend's large demo grids.
- **`-rows`/`-cols` are both required** when running this binary directly (unlike `bil emu`'s `-rows`/`-cols`, `0` doesn't mean "infer" here) — this backend's grids are always small, so there's little to infer, and a too-small grid can silently under-exercise a program the same way it can for `bil emu` (e.g. `processor(0,0)` and `processor(0,cols-1)` colliding into the same match when `cols=1`) — size it deliberately. `bil emu -target multicore` *can* infer a size (it reuses the same inference `-target wasm` does, then checks the result against this machine's core count), so this only bites when running `cmd/multicore` standalone.
- **No web page, no visual grid.** Each process's `Screenf` output goes straight to this backend's own stdout, prefixed `[row,col]`; a crashed node's stderr (a Go panic, not just a `Screenf` call) gets the same prefix treatment so it's still attributable.
- **No in-page demo selector, no live resize.** Every run builds fresh from `-dir` and exits when the grid's node programs do (or on Ctrl-C) — there's no persistent server to restart against a different binary or grid size, unlike `cmd/wasm`.
- **Sockets, not shared memory.** Unlike the browser backend's `SharedArrayBuffer`, each wired link is a pair of Unix-domain-socket connections in a per-run temp directory — cleaned up on exit, but a killed-with-`-9` run can leave stale socket files behind (harmless; a fresh run's listener side removes any stale file at its own path before binding).
