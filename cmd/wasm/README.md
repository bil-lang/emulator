# cmd/wasm — the browser backend

Each grid node runs as an isolated Go-compiled-to-WASM program inside its own dedicated Web Worker — a real OS thread with a real isolated linear memory per node — wired to its N/E/S/W nearest neighbours by bilateral serial links, each a genuine blocking rendezvous (see [`bilink`](../../bilink)), for real hardware-realistic isolation, not a goroutines-and-Go-channels simulation. `main.go` here is just a minimal static file server (`static/`, alongside this file) that sets the two response headers `SharedArrayBuffer` requires; `static/index.html` and `static/node-worker.js` are where the actual grid — topology, Worker spawning, the link primitive's JS half — lives.

See the top-level [README.md](../../README.md) for how this fits alongside [`cmd/multicore`](../multicore), and for the shared list of node programs.

## Quickstart: run a placement example

This walks through `bil/examples/20-placed-controller.bil` (Phase C, the first heterogeneous placement demo) end to end, from the `emulator` repo root. Swap the example/nodeprog name for any of the others in the top-level README's "Node programs" list.

```
# 1. build bilc
(cd ../bil/tools/bilc/cmd/bilc && go build -o ../../bilc .)

# 2. transpile the example -- writes nodeprog/placedcontroller/{main.go, roles/}
mkdir -p nodeprog/placedcontroller
../bil/tools/bilc/bilc ../bil/examples/20-placed-controller.bil nodeprog/placedcontroller/main.go

# 3. build it as the one binary every node runs (bilc's own runtime switch picks the role)
GOOS=js GOARCH=wasm go build -o cmd/wasm/static/node.wasm ./nodeprog/placedcontroller

# 4. serve and open it
go run ./cmd/wasm -addr localhost:8789
```

Open `http://localhost:8789/` — `SharedArrayBuffer` requires the COOP/COEP headers this server sets, so this won't work opened as a bare `file://` page, or served some other way without those headers. It auto-starts a 6×7 grid; the rows/cols fields and "Restart grid" button let you try other sizes without rebuilding anything.

**To switch demos**, redo steps 2-3 for a different example/`nodeprog/<name>`, overwriting `cmd/wasm/static/node.wasm`, then refresh the page (or click "Restart grid") — there's no in-page selector; whichever binary is currently at `cmd/wasm/static/node.wasm` is what every node runs. Step 4's server doesn't need restarting.

Step 3 above is the simplest path (one binary, `bilc`'s own runtime `switch` dispatches each node to its role). For the real link-native boot cascade instead — separate per-role binaries, hop-relayed rather than told directly — see "Host and boot cascade" below. Or skip all of this and run `bil emu examples/20-placed-controller.bil` from the sibling `bil` repo, which does steps 1-4 itself.

## Host and boot cascade

For any `placed par` program, `bilc` also emits `roles/<name>/main.go` — one standalone Go source per distinct role — and `roles/placement.json`, describing every reachable leaf's clause match, `if`/`else` condition chain, and link-index bindings (see `../../../bil/tools/bilc/bilc.go`'s `RoleBinaries`/`PlacementManifest`). Build each role and copy the manifest into `static/roles/` (also gitignored — built fresh alongside the `nodeprog/` source each time):

```
for role in controller rowEnd relay idle; do
  GOOS=js GOARCH=wasm go build -o cmd/wasm/static/roles/$role.wasm ./nodeprog/placedcontroller/roles/$role
done
cp nodeprog/placedcontroller/roles/placement.json cmd/wasm/static/roles/placement.json
```

When `static/roles/placement.json` exists, `index.html`'s `startGrid` uses a genuinely link-native bootstrap instead of the direct-postMessage one every other demo above still uses: a **host** (this same page, acting as the one external entry point) injects role, identity, and the actual compiled bytes only at processor (0,0), and every other node learns all three solely by relaying hop-by-hop across a spanning tree of the grid (`bootChildCoords`) — never told directly by the page the way `startGridLegacy` tells every Worker outright. This fetches each distinct role's `.wasm` exactly once (4 requests for a 6×7 `placedcontroller` grid, not 42), fanning it out afterward via `MessageChannel`s rather than redundant network fetches — real transputer hardware has no side-channel to every node either, only nearest-neighbour links, and this is the same shape: `Row()`/`Col()`/etc. end up set identically either path (see `bilink.go`), so a node program itself can't tell which bootstrap ran.

The actual hop-by-hop relay uses `MessageChannel`s, not `bilink`'s own `SharedArrayBuffer`+`Atomics` link primitive — deliberately: that primitive moves one 32-bit word per round trip, fine for ordinary Bil channel traffic but far too slow to relay multi-megabyte binaries through. `roles/placement.json`'s `transport` field names which boot-cascade mechanism a launch expects (`message-channel` is the only one implemented; `link-chunked` — a real, slower, chunked-over-`link[0]` transport for a genuinely link-only, non-browser target — is named but not built).

## Quirks

- **Needs COOP/COEP headers, always.** `SharedArrayBuffer` doesn't exist in the page at all without them — a bare `file://` open, or serving `static/` with any other tool that doesn't set `Cross-Origin-Opener-Policy: same-origin` + `Cross-Origin-Embedder-Policy: require-corp`, silently fails. Always go through this backend's own `main.go`.
- **No in-page demo selector.** Whichever binary currently sits at `static/node.wasm` (or `static/roles/*.wasm` + `placement.json`) is what runs — switching demos means rebuilding and refreshing, not picking from a menu.
- **Spawning is throttled.** `index.html` spawns Workers in small batches (`SPAWN_BATCH_SIZE = 6`, `SPAWN_BATCH_DELAY_MS = 200`) rather than all at once — each Worker independently fetches and compiles a ~2.6MB WASM module and boots a full separate Go runtime (own scheduler, GC, heap), and firing 100+ of those concurrently silently stalls some of them with no error at all (confirmed even at 12×12 = 144 nodes on a freshly loaded page).
- **Grid size is a runtime choice, not a build one.** The rows/cols fields and "Restart grid" button resize without rebuilding — grid size lives entirely in `index.html`'s own state, never baked into the `.wasm`.
- **Binary size doesn't shrink per role.** Measured directly: splitting into per-role binaries barely changes each `.wasm`'s own size (~2.66MB full-switch vs. ~2.65–2.66MB per role) — Go's own runtime (scheduler, GC, `reflect` for `altN`) dominates the binary, not the actual role code. Per-role splitting fixes redundant network fetches at grid scale, not per-node memory footprint. TinyGo (not yet evaluated) is the more likely lever for binary size itself.
- **One-Worker-is-one-node, always.** A node program must not try to enumerate the grid itself (no `par`/`par range` in a `.bil` node program, no spawning N×M goroutines in a hand-written one) — this backend already spawns and positions every node's own Worker; a node program's only job is to run once, discovering who it is via `bilink.Row()`/`Col()`. Ignoring this produced a real bug during Phase A→B verification: a nested replicated `par` in `main()` spawned 42 goroutines inside one Worker, all reading that one Worker's fixed identity and racing on the same link buffers.
