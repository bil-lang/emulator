# emulator

A genuine (not simulated) software emulation environment for an NxM array of processors, built for the [Bil](../bil) language project. Each grid node runs as an isolated Go-compiled-to-WASM program inside its own dedicated Web Worker — a real OS thread with a real isolated linear memory per node — wired to its N/E/S/W nearest neighbours by point-to-point serial links, each a genuine blocking rendezvous over a `SharedArrayBuffer`. See `../bil/STRATEGY.md`'s "Emulator architecture" note for why this is built this way rather than as a goroutines-and-Go-channels simulation.

## Running it

```
go run ./cmd/serve -dir static -addr localhost:8789
```

Then open `http://localhost:8789/` — `SharedArrayBuffer` requires the COOP/COEP headers `cmd/serve` sets, so this won't work opened as a bare `file://` page. It auto-starts a 6×7 grid running whatever is currently built as `static/node.wasm`; the rows/cols fields and "Restart grid" button let you try other sizes.

## Node programs

- `nodeprog/ripple` — Phase A: a hand-written Go demo proving the substrate (isolation + real blocking links) works at all, with no dependency on `bil`.
- `nodeprog/meshripple` — Phase B: the same ripple demo, but compiled from a real `.bil` program (`../bil/examples/19-mesh-ripple.bil`) via `../bil/tools/bilc`'s `placed par`/`link[i].in`/`.out` support — requires a sibling checkout of the `bil` repo at `../bil`. Checked in as generated Go (see its header for the regenerate command) rather than requiring a build step to try it, alongside `roles/deploy.json`. Three roles wildcarded across every row — `origin` at column 0, `reflect` at the last column, `relay` everywhere else — the direct ancestor of `nodeprog/placedcontroller`'s own controller/rowEnd/relay split below.
- `nodeprog/placedcontroller` — Phase C: a genuinely *heterogeneous* demo, compiled from `../bil/examples/20-placed-controller.bil` via `bilc`'s `placed par`/`processor(...)`/`place ... at link[...]` support — four separately-named, separately-reasoned-about procs (`controller`, `rowEnd`, `relay`, `idle`), each dispatched to specific processors by the `.bil` source itself rather than by an `if`/`else` inside one shared proc. Also checked in as generated Go, alongside `roles/deploy.json` — a complete, mechanical description of the declared placement, generated for free from the same `.bil` source (see "Host and boot cascade" below).
- `nodeprog/meshtransformer` — Phase D: a minimal LLM-style transformer block, compiled from `../bil/examples/21-mesh-transformer.bil`. Reuses placedcontroller's exact controller/rowEnd/relay/idle placement idiom, but generalizes the single reflected counter into gathering a whole row's input tokens to the controller and scattering back that many predicted tokens — real embeddings and attention scores crossing genuine cross-Worker links, one int32 at a time, feeding a real (if tiny and untrained) embedding + positional-encoding + self-attention + feed-forward + output-projection pass that runs on the controller once the whole row has arrived.
- `nodeprog/pipelinetransformer` — Phase E: the same tiny transformer, but this time the mesh's own two dimensions do the actual compute distribution instead of just relaying data to one processor. Chunk by column, pipeline by row: each column is one token position, and each of the grid's first 4 rows is one transformer sub-stage (embed+posenc, attention+residual, feed-forward+residual, output-proj+argmax), with a token's representation flowing south from stage to stage. Three of the four stages are purely per-token, so every column in that row computes in parallel on its own Worker; only attention is inherently cross-token, so row 1 alone falls back to placedcontroller's gather/relay/scatter idiom, now carrying a whole `[dModel]float64` vector per hop (as 4 float32-bit-encoded int32s) instead of one scalar. Because links are unbuffered rendezvous, generations pipeline across the 4 stages for free — row 0 can start the next generation's embedding as soon as row 1 has drained the current one, without waiting for row 2/3 to finish it — so throughput is gated only by the slowest stage (attention), the same way a hardware pipeline's throughput is gated by its slowest stage. Compiled from `../bil/examples/22-pipeline-transformer.bil`; needs at least 4 mesh rows (the default 6×7 grid qualifies).

Any of these needs building for `static/node.wasm` before `cmd/serve` will show it:

```
GOOS=js GOARCH=wasm go build -o static/node.wasm ./nodeprog/ripple      # or ./nodeprog/meshripple, ./nodeprog/placedcontroller, ./nodeprog/meshtransformer, ./nodeprog/pipelinetransformer
```

## Host and boot cascade

For any `placed par` program (`placedcontroller` today), `bilc` also emits `roles/<name>/main.go` — one standalone Go source per distinct role — and `roles/deploy.json`, describing every reachable leaf's clause match, `if`/`else` condition chain, and link-index bindings (see `../bil/tools/bilc/bilc.go`'s `RoleBinaries`/`DeployManifest`). Build each role and copy the manifest into `static/`:

```
for role in controller rowEnd relay idle; do
  GOOS=js GOARCH=wasm go build -o static/roles/$role.wasm ./nodeprog/placedcontroller/roles/$role
done
cp nodeprog/placedcontroller/roles/deploy.json static/roles/deploy.json
```

When `static/roles/deploy.json` exists, `index.html`'s `startGrid` uses a genuinely link-native bootstrap instead of the direct-postMessage one every other demo above still uses: a **host** (this same page, acting as the one external entry point) injects role, identity, and the actual compiled bytes only at processor (0,0), and every other node learns all three solely by relaying hop-by-hop across a spanning tree of the grid (`bootChildCoords`) — never told directly by the page the way `startGridLegacy` tells every Worker outright. This fetches each distinct role's `.wasm` exactly once (4 requests for a 6×7 `placedcontroller` grid, not 42), fanning it out afterward via `MessageChannel`s rather than redundant network fetches — real transputer hardware has no side-channel to every node either, only nearest-neighbour links, and this is the same shape: `Row()`/`Col()`/etc. end up set identically either path (see `bilink.go`), so a node program itself can't tell which bootstrap ran.

The actual hop-by-hop relay uses `MessageChannel`s, not `bilink`'s own `SharedArrayBuffer`+`Atomics` link primitive — deliberately: that primitive moves one 32-bit word per round trip, fine for ordinary Bil channel traffic but far too slow to relay multi-megabyte binaries through. `roles/deploy.json`'s `transport` field names which boot-cascade mechanism a launch expects (`message-channel` is the only one implemented; `link-chunked` — a real, slower, chunked-over-`link[0]` transport for a genuinely link-only, non-browser target — is named but not built).

Measured directly: splitting into per-role binaries barely changes each `.wasm`'s own size (~2.66MB full-switch vs. ~2.65–2.66MB per role) — Go's own runtime (scheduler, GC, `reflect` for `altN`) dominates the binary, not the actual role code, so this doesn't fix per-node memory footprint. What it does fix is redundant network fetches at grid scale. TinyGo (not yet evaluated) is the more likely lever for binary size itself.

**One-Worker-is-one-node, always:** a node program must not try to enumerate the grid itself (no `par`/`par range` in a `.bil` node program, no spawning N*M goroutines in a hand-written one) — the emulator already spawns and positions every node's own Worker; a node program's only job is to run once, discovering who it is via `bilink.Row()`/`Col()`. Ignoring this produced a real bug during Phase A→B verification: a nested replicated `par` in `main()` spawned 42 goroutines inside one Worker, all reading that one Worker's fixed identity and racing on the same link buffers.
