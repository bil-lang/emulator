# emulator

A genuine (not simulated) software emulation environment for an NxM array of processors, built for the [Bil](../bil) language project. Each grid node runs as an isolated Go-compiled-to-WASM program inside its own dedicated Web Worker — a real OS thread with a real isolated linear memory per node — wired to its N/E/S/W nearest neighbours by point-to-point serial links, each a genuine blocking rendezvous over a `SharedArrayBuffer`. See `../bil/STRATEGY.md`'s "Emulator architecture" note for why this is built this way rather than as a goroutines-and-Go-channels simulation.

## Running it

```
go run ./cmd/serve -dir static -addr localhost:8789
```

Then open `http://localhost:8789/` — `SharedArrayBuffer` requires the COOP/COEP headers `cmd/serve` sets, so this won't work opened as a bare `file://` page. It auto-starts a 6×7 grid running whatever is currently built as `static/node.wasm`; the rows/cols fields and "Restart grid" button let you try other sizes.

## Node programs

- `nodeprog/ripple` — Phase A: a hand-written Go demo proving the substrate (isolation + real blocking links) works at all, with no dependency on `bil`.
- `nodeprog/meshripple` — Phase B: the same ripple demo, but compiled from a real `.bil` program (`../bil/examples/18-mesh-ripple.bil`) via `../bil/tools/bilc`'s `link[i]` support — requires a sibling checkout of the `bil` repo at `../bil`. Checked in as generated Go (see its header for the regenerate command) rather than requiring a build step to try it.

Either one needs building for `static/node.wasm` before `cmd/serve` will show it:

```
GOOS=js GOARCH=wasm go build -o static/node.wasm ./nodeprog/ripple      # or ./nodeprog/meshripple
```

**One-Worker-is-one-node, always:** a node program must not try to enumerate the grid itself (no `par`/`par range` in a `.bil` node program, no spawning N*M goroutines in a hand-written one) — the emulator already spawns and positions every node's own Worker; a node program's only job is to run once, discovering who it is via `bilink.Row()`/`Col()`. Ignoring this produced a real bug during Phase A→B verification: a nested replicated `par` in `main()` spawned 42 goroutines inside one Worker, all reading that one Worker's fixed identity and racing on the same link buffers.
