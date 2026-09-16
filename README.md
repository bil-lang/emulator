# emulator

A software emulation environment for an NxM array of processors, built for the [Bil](../bil) language project. Two backends run the exact same node-program code over the exact same [`bilink`](bilink) link/channel primitive, but give it two different kinds of real hardware isolation:

| | [`cmd/wasm`](cmd/wasm) | [`cmd/multicore`](cmd/multicore) |
|---|---|---|
| Execution unit | one browser Web Worker per grid node | one native OS process per grid node |
| Grid size | large (demoed at 6×7 = 42 nodes) | capped at this machine's CPU core count |
| Core mapping | none — a Worker, not a pinned core | real pin on Linux; scheduler-only on macOS |
| Direct launch | `go run ./cmd/wasm`, opens a browser tab | `go run ./cmd/multicore`, plain terminal output |
| Via `bil emu` | `bil emu file.bil` (default) | `bil emu -target multicore file.bil` |
| Quickstart & quirks | [cmd/wasm/README.md](cmd/wasm/README.md) | [cmd/multicore/README.md](cmd/multicore/README.md) |

**This repo only matters for parallel Bil that targets a physical mesh** — specifically programs using `link[...]`, `place`, or `placed par` (`bil/examples/19` through `23` today). Otherwise Bil's examples are ordinary `chan`/`proc`/`par` programs, run directly with `bil run examples/N.bil` concurrently on a normal machine.

## Install

1. A Go toolchain on `PATH` (same requirement as `bil` itself). Neither backend needs cgo — `cmd/multicore`'s real core pin on Linux is a plain syscall, not cgo (see its own README).
2. A sibling checkout of [`bil`](../bil) at `../bil` — this repo builds node programs by running `bil`'s own `bilc` transpiler against `bil/examples/*.bil`, and needs that path to resolve.
3. Clone this repo itself as `../emulator` next to `bil`. It has no public GitHub remote yet (still local-only) — for now, just have both directories side by side; once it's pushed to `github.com/bil-lang/emulator`, this step becomes a normal `git clone`.

## bilink

[`bilink`](bilink) is the one shared substrate every node program calls into — `Send`/`Recv`/`Row`/`Col`/`NumRows`/`NumCols`/`NumLinks`/`ID`/`LinkWired`/`Screenf` — and the only thing that actually differs per backend. It has two implementations, picked automatically by Go's own build-tag resolution depending on what you build a node program for:

- `bilink.go` (`js && wasm`) — backed by `syscall/js` + `SharedArrayBuffer`/`Atomics`, for `cmd/wasm`.
- `bilink_native.go` + `affinity_linux.go`/`affinity_stub.go` (`!js`) — backed by Unix-domain sockets (+ a real core pin on Linux), for `cmd/multicore`.

A node program itself never needs to know or care which backend it ends up running under; both implementations provide the exact same genuine, unbuffered, two-sided rendezvous, just over a different real transport. See each backend's own README for the details of its transport.

## Node programs

- `nodeprog/ripple` — Phase A: a hand-written Go demo proving the substrate (isolation + real blocking links) works at all, with no dependency on `bil`. Checked in; nothing to build from source for this one.
- `nodeprog/meshripple` — Phase B: the same ripple demo, but compiled from a real `.bil` program (`../bil/examples/10-mesh-ripple.bil`) via `../bil/tools/bilc`'s `placed par`/`link[i].in`/`.out` support. Three roles wildcarded across every row — `origin` at column 0, `reflect` at the last column, `relay` everywhere else — the direct ancestor of `nodeprog/placedcontroller`'s own controller/rowEnd/relay split below.
- `nodeprog/placedcontroller` — Phase C: a genuinely *heterogeneous* demo, compiled from `../bil/examples/20-placed-controller.bil` via `bilc`'s `placed par`/`processor(...)`/`place ... at link[...]` support — four separately-named, separately-reasoned-about procs (`controller`, `rowEnd`, `relay`, `idle`), each dispatched to specific processors by the `.bil` source itself rather than by an `if`/`else` inside one shared proc.
- `nodeprog/meshtransformer` — Phase D: a minimal LLM-style transformer block, compiled from `../bil/examples/21-mesh-transformer.bil`. Reuses placedcontroller's exact controller/rowEnd/relay/idle placement idiom, but generalizes the single reflected counter into gathering a whole row's input tokens to the controller and scattering back that many predicted tokens — real embeddings and attention scores crossing genuine cross-node links, one int32 at a time, feeding a real (if tiny and untrained) embedding + positional-encoding + self-attention + feed-forward + output-projection pass that runs on the controller once the whole row has arrived.
- `nodeprog/pipelinetransformer` — Phase E: the same tiny transformer, but this time the mesh's own two dimensions do the actual compute distribution instead of just relaying data to one processor. Chunk by column, pipeline by row: each column is one token position, and each of the grid's first 4 rows is one transformer sub-stage (embed+posenc, attention+residual, feed-forward+residual, output-proj+argmax), with a token's representation flowing south from stage to stage. Three of the four stages are purely per-token, so every column in that row computes in parallel on its own node; only attention is inherently cross-token, so row 1 alone falls back to placedcontroller's gather/relay/scatter idiom, now carrying a whole `[dModel]float64` vector per hop (as 4 float32-bit-encoded int32s) instead of one scalar. Because links are unbuffered rendezvous, generations pipeline across the 4 stages for free — row 0 can start the next generation's embedding as soon as row 1 has drained the current one, without waiting for row 2/3 to finish it — so throughput is gated only by the slowest stage (attention), the same way a hardware pipeline's throughput is gated by its slowest stage. Needs at least 4 mesh rows.
- `nodeprog/corebench` — a LINPACK-style CPU-bound benchmark, compiled from `../bil/examples/23-corebench.bil`. Placement is a single row, wildcarded (`processor(0, *)`), so the same compiled roles run unmodified at any grid width — built to make `cmd/multicore`'s real core usage visible (see its README), though it also runs fine on `cmd/wasm` like every other demo here.

Every demo except `ripple` is built fresh from `bil/examples/*.bil`, not checked in (`nodeprog/{meshripple,placedcontroller,meshtransformer,pipelinetransformer,corebench}/` are gitignored) — see either backend's README for the exact build commands; swap `placedcontroller`/`20-placed-controller.bil` for `meshripple`/`10-mesh-ripple.bil`, `meshtransformer`/`21-mesh-transformer.bil`, `pipelinetransformer`/`22-pipeline-transformer.bil`, or `corebench`/`23-corebench.bil` to try a different one.

## `bil emu -target`

The sibling `bil` repo's `bil emu <file.bil>` command (`bil/tools/bil/emu.go`) transpiles and role-splits a `.bil` program exactly as the Node programs section above describes, then hands the result to whichever backend `-target` names — `wasm` (the default, matching `bil emu`'s original behavior) or `multicore`:

```
bil emu examples/20-placed-controller.bil                    # cmd/wasm, browser
bil emu -target multicore examples/20-placed-controller.bil  # cmd/multicore, native
```

This is the one-command path; either backend's own README below documents the equivalent manual steps (`bilc` + the backend's launcher) for when you want to inspect or tweak what's actually being built.
