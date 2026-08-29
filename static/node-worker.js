// node-worker.js: bootstraps one grid node's isolated Go/WASM instance
// inside its own dedicated Worker -- a real OS thread, a real isolated
// linear memory per node. Implements the link primitive using
// Atomics.waitAsync only -- never a synchronous Atomics.wait, which
// would freeze every goroutine in this node's WASM instance (see
// ../../bil/STRATEGY.md's "Emulator architecture" note, and the
// ../spike/ that proved this mechanism).
//
// Links are addressed by plain index, not compass-direction names --
// this 2D-mesh topology's own convention is slot 0=north, 1=east,
// 2=south, 3=west (see static/index.html's buildTopology), but that
// mapping lives entirely here and in the page; node programs (and the
// bilink Go package) just see link[i] and don't know or care what
// topology assigned it. A boundary node simply has some slots absent.

const STATE = 0; // int32 index: 0=idle, 1=data-ready, 2=consumed
const PAYLOAD = 1; // int32 index: the transferred value
const IDLE = 0, DATA_READY = 1, CONSUMED = 2;

// Go panics/fatal errors print via console.error (wasm_exec.js's
// stderr shim), not as an uncaught exception this Worker's onerror
// would catch -- so without this, a crashed node just goes silent
// forever instead of visibly failing. Forward it to the main thread.
const _consoleError = console.error.bind(console);
console.error = (...args) => {
  _consoleError(...args);
  postMessage({ type: "error", row: self.bilRow, col: self.bilCol, text: args.map(String).join(" ") });
};
self.onerror = (e) => {
  postMessage({ type: "error", row: self.bilRow, col: self.bilCol, text: e.message });
};

let views = []; // views[i] = {out, in} of Int32Array, or undefined if slot i isn't wired

// Atomics.waitAsync only guarantees "the value changed since it was
// last known to equal the value passed in" -- not that it became any
// specific value -- so this re-checks on every wake and retries if
// it's a spurious/unrelated change.
function waitUntil(view, idx, target, cb) {
  function attempt() {
    const cur = Atomics.load(view, idx);
    if (cur === target) {
      cb();
      return;
    }
    const w = Atomics.waitAsync(view, idx, cur);
    if (!w.async) {
      attempt();
      return;
    }
    w.value.then(attempt);
  }
  attempt();
}

function requireLink(i) {
  const link = views[i];
  if (!link) {
    throw new Error(
      `node (${self.bilRow},${self.bilCol}) has no link[${i}] wired -- ` +
      `check the node program is using a valid index for its position`
    );
  }
  return link;
}

self.linkSend = function (i, value, cb) {
  const view = requireLink(i).out;
  waitUntil(view, STATE, IDLE, () => {
    Atomics.store(view, PAYLOAD, value);
    Atomics.store(view, STATE, DATA_READY);
    Atomics.notify(view, STATE);
    waitUntil(view, STATE, CONSUMED, () => {
      Atomics.store(view, STATE, IDLE);
      Atomics.notify(view, STATE);
      cb();
    });
  });
};

self.linkRecv = function (i, cb) {
  const view = requireLink(i).in;
  waitUntil(view, STATE, DATA_READY, () => {
    const v = Atomics.load(view, PAYLOAD);
    Atomics.store(view, STATE, CONSUMED);
    Atomics.notify(view, STATE);
    cb(v);
  });
};

self.screenPrint = function (text) {
  postMessage({ type: "screen", row: self.bilRow, col: self.bilCol, text });
};

self.onmessage = (e) => {
  const msg = e.data;
  self.bilRow = msg.row;
  self.bilCol = msg.col;
  self.bilRows = msg.rows;
  self.bilCols = msg.cols;
  self.bilNumLinks = msg.links.length;

  views = msg.links.map((link) =>
    link ? { out: new Int32Array(link.out), in: new Int32Array(link.in) } : undefined
  );

  importScripts("wasm_exec.js");
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("node.wasm"), go.importObject).then((r) => {
    go.run(r.instance);
  });
};
