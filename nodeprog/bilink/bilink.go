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
//go:build js && wasm

package bilink

import (
	"fmt"
	"syscall/js"
)

// Direction identifies one of a node's up to 4 nearest-neighbour
// links. A node at a grid edge simply has no wiring for the missing
// directions; calling Send/Recv on one that isn't wired is a bug in
// the calling node program, not something this package guards against.
type Direction int

const (
	North Direction = iota
	East
	South
	West
)

func (d Direction) String() string {
	switch d {
	case North:
		return "north"
	case East:
		return "east"
	case South:
		return "south"
	case West:
		return "west"
	default:
		return "?"
	}
}

// Send blocks until the value has been both delivered to, and
// acknowledged as consumed by, the neighbour in direction d -- a true
// two-sided rendezvous, matching Bil/occam's "channels never buffer"
// rule (see bil/LANG-DESIGN.md and bil/intro-to-bil.html Rule 01).
func Send(d Direction, v int32) {
	done := make(chan struct{})
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) any {
		cb.Release()
		close(done)
		return nil
	})
	js.Global().Call("linkSend", d.String(), v, cb)
	<-done
}

// Recv blocks until a value arrives from the neighbour in direction d.
func Recv(d Direction) int32 {
	result := make(chan int32, 1)
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) any {
		cb.Release()
		result <- int32(args[0].Int())
		return nil
	})
	js.Global().Call("linkRecv", d.String(), cb)
	return <-result
}

// Row, Col, NumRows, NumCols report this node's position and the
// overall grid size, set by the Worker bootstrap before the WASM
// program starts running.
func Row() int      { return js.Global().Get("bilRow").Int() }
func Col() int      { return js.Global().Get("bilCol").Int() }
func NumRows() int  { return js.Global().Get("bilRows").Int() }
func NumCols() int  { return js.Global().Get("bilCols").Int() }

// Screenf writes one line to this node's on-page mini screen.
func Screenf(format string, a ...any) {
	js.Global().Call("screenPrint", fmt.Sprintf(format, a...))
}
