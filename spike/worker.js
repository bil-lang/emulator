// Throwaway spike worker. Receives two SharedArrayBuffers from the
// main thread, then runs a Go/WASM program that blocks on
// "waitForChange" TWICE in sequence -- implemented here via
// Atomics.waitAsync, never a synchronous Atomics.wait, so neither
// wait ever freezes the whole WASM instance (see spike/main.go).

let views = [];

function waitForChange(which, cb) {
  const view = views[which];
  const idx = 0;
  function attempt() {
    const cur = Atomics.load(view, idx);
    if (cur !== 0) {
      cb();
      return;
    }
    const w = Atomics.waitAsync(view, idx, 0);
    if (!w.async) {
      attempt();
      return;
    }
    w.value.then(() => attempt());
  }
  attempt();
}
self.waitForChange = waitForChange;

self.onmessage = (e) => {
  views = e.data.sabs.map((sab) => new Int32Array(sab));
  importScripts("wasm_exec.js");
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((r) => {
    go.run(r.instance);
  });
};
