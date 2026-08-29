// Package bilink is the Go-side API for the emulator's link primitive:
// a genuine blocking rendezvous between adjacent grid nodes. On the JS
// side (node-worker.js) this is implemented with Atomics.waitAsync --
// never a synchronous Atomics.wait, which would freeze every goroutine
// in this node's WASM instance, since WASM has one call stack per
// instance and no preemption. Send/Recv here each park only the one
// calling goroutine, via an ordinary Go channel receive, so Go's own
// scheduler keeps running everything else in this node while a proc
// waits on a link -- see ../../README.md and ../../../bil/STRATEGY.md's
// "Emulator architecture" note for why this matters.
//
// Links are addressed by plain index (link[0], link[1], ...) rather
// than a fixed set of compass-direction names, so this package makes
// no assumption about topology -- a 2D mesh (4 links), a 3D mesh (6),
// or anything else all just pick however many slots they need. What
// index means what physical neighbour is a convention owned by the
// topology/layout layer (see node-worker.js and static/index.html),
// not by this package or by a node program.
//
//go:build js && wasm

package bilink

import (
	"fmt"
	"syscall/js"
)

// Send blocks until v has been both delivered to, and acknowledged as
// consumed by, whatever this node is wired to via link[idx] -- a true
// two-sided rendezvous, matching Bil/occam's "channels never buffer"
// rule (see bil/LANG-DESIGN.md and bil/intro-to-bil.html Rule 01).
// Calling Send on an idx this node has no wiring for is a bug in the
// calling node program, not something this package guards against.
func Send(idx int, v int32) {
	done := make(chan struct{})
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) any {
		cb.Release()
		close(done)
		return nil
	})
	js.Global().Call("linkSend", idx, v, cb)
	<-done
}

// Recv blocks until a value arrives via link[idx].
func Recv(idx int) int32 {
	result := make(chan int32, 1)
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) any {
		cb.Release()
		result <- int32(args[0].Int())
		return nil
	})
	js.Global().Call("linkRecv", idx, cb)
	return <-result
}

// Row, Col, NumRows, NumCols report this node's position and the
// overall grid size. NumLinks reports how many link slots this
// topology defines -- link[0]..link[NumLinks()-1] are the valid
// indices, though a node at a boundary may not have every slot
// actually wired to a neighbour. All are set by the Worker bootstrap
// before the WASM program starts running.
func Row() int      { return js.Global().Get("bilRow").Int() }
func Col() int      { return js.Global().Get("bilCol").Int() }
func NumRows() int  { return js.Global().Get("bilRows").Int() }
func NumCols() int  { return js.Global().Get("bilCols").Int() }
func NumLinks() int { return js.Global().Get("bilNumLinks").Int() }

// Screenf writes one line to this node's on-page mini screen.
func Screenf(format string, a ...any) {
	js.Global().Call("screenPrint", fmt.Sprintf(format, a...))
}
