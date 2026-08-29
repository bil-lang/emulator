// Throwaway spike worker. Receives a SharedArrayBuffer from the main
// thread, then runs a Go/WASM program that blocks one goroutine on
// "waitForChange" -- implemented here via Atomics.waitAsync, never a
// synchronous Atomics.wait, so this goroutine's wait never freezes the
// whole WASM instance (see spike/main.go and ../../bil/STRATEGY.md).

let linkView = null;

// Atomics.waitAsync only guarantees "the value changed since it was
// last known to equal `expected`" -- not that it became any specific
// value. So this must re-check the real value on every wake and retry
// if it's a spurious/timeout wake that didn't actually change it.
function waitForChange(cb) {
  const idx = 0;
  function attempt() {
    const cur = Atomics.load(linkView, idx);
    if (cur !== 0) {
      cb();
      return;
    }
    const w = Atomics.waitAsync(linkView, idx, 0);
    if (!w.async) {
      // value already changed between the load above and this call
      attempt();
      return;
    }
    w.value.then(() => attempt());
  }
  attempt();
}
self.waitForChange = waitForChange;

self.onmessage = (e) => {
  linkView = new Int32Array(e.data.sab);
  importScripts("wasm_exec.js");
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((r) => {
    go.run(r.instance);
  });
};
