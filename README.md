# emulator

A genuine (not simulated) software emulation environment for an NxM array of processors, built for the [Bil](../bil) language project. Each grid node runs as an isolated Go-compiled-to-WASM program inside its own dedicated Web Worker — a real OS thread with a real isolated linear memory per node — wired to its N/E/S/W nearest neighbours by point-to-point serial links, each a genuine blocking rendezvous over a `SharedArrayBuffer`. See `../bil/STRATEGY.md`'s "Emulator architecture" note for why this is built this way rather than as a goroutines-and-Go-channels simulation.

This repo is self-contained for its first milestone (a hand-written demo program proving the substrate works). A later milestone compiles real `.bil` programs onto the grid via `../bil/tools/bilc`, which requires a sibling checkout of the `bil` repo at `../bil` relative to this one.
