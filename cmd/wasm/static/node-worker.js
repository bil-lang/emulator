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

// A crashed/panicking node's failure otherwise surfaces only as a
// console.error in this Worker's own separate DevTools context, not as
// an uncaught exception this Worker's onerror would catch -- so without
// this, a crashed node just goes silent forever instead of visibly
// failing. Forward it to the main thread.
const _consoleError = console.error.bind(console);
console.error = (...args) => {
  _consoleError(...args);
  postMessage({ type: "error", row: self.bilRow, col: self.bilCol, text: args.map(String).join(" ") });
};
self.onerror = (e) => {
  postMessage({ type: "error", row: self.bilRow, col: self.bilCol, text: e.message });
};

// wasm_exec.js's own runtime.wasmWrite -> fs.writeSync always calls
// console.log for every stdout/stderr write -- it ignores the file
// descriptor entirely, so this is what plain `println`/`fmt.Println`
// (not just an explicit bilink.Screenf call, which posts to the main
// thread directly via screenPrint below) come out through. Wrapping it
// here, before wasm_exec.js is even loaded, means any node program's
// ordinary output shows up on its own page tile -- not just programs
// that call Screenf.
const _consoleLog = console.log.bind(console);
console.log = (...args) => {
  _consoleLog(...args);
  postMessage({ type: "screen", row: self.bilRow, col: self.bilCol, text: args.map(String).join(" ") });
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

// linkWired lets a running node program check whether link[i] actually
// has a neighbour wired up before using it, instead of only finding out
// via requireLink's own throw on misuse -- see bilink.LinkWired's own
// doc comment for why placement makes this worth having (several
// differently-roled procs, each assuming their own subset of links).
self.linkWired = function (i) {
  return views[i] !== undefined;
};

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

  views = msg.links.map((link) =>
    link ? { out: new Int32Array(link.out), in: new Int32Array(link.in) } : undefined
  );

  if (msg.bootParentPort) {
    // Link-native boot cascade (see index.html's startGridCascade):
    // this node's role, its identity (row/col/rows/cols), and the
    // actual WASM bytes to run arrive only via this port -- relayed
    // hop-by-hop from a host attached at (0,0), never told directly by
    // the page the way the legacy bootstrap below does. bilRow/bilCol
    // etc. end up set exactly the same way either path, so bilink.go's
    // own Row()/Col() accessors need no knowledge of which path ran.
    msg.bootParentPort.onmessage = (be) => {
      const payload = be.data;
      self.bilRow = payload.identity.row;
      self.bilCol = payload.identity.col;
      self.bilRows = payload.identity.rows;
      self.bilCols = payload.identity.cols;
      self.bilNumLinks = payload.identity.numLinks;

      for (const item of payload.forward) {
        msg.bootChildPorts[item.childIndex].postMessage(item.payload);
      }

      importScripts("wasm_exec.js");
      const go = new Go();
      WebAssembly.instantiate(payload.wasmBytes, go.importObject).then((r) => {
        go.run(r.instance);
      });
    };
    return;
  }

  // Legacy bootstrap: role/identity told directly by the page, one
  // shared node.wasm fetched by every Worker itself -- still used by
  // any program with no roles/placement.json (see index.html's startGrid).
  self.bilRow = msg.row;
  self.bilCol = msg.col;
  self.bilRows = msg.rows;
  self.bilCols = msg.cols;
  self.bilNumLinks = msg.links.length;

  importScripts("wasm_exec.js");
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("node.wasm"), go.importObject).then((r) => {
    go.run(r.instance);
  });
};
