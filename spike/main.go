// Throwaway spike: proves that blocking on the async link primitive
// (Atomics.waitAsync -> js.FuncOf callback -> Go channel receive) only
// parks the one goroutine doing it, and does not freeze the rest of
// this node's Go program. If this holds, Go's own scheduler is enough
// to give each node real occam-style concurrency -- no bespoke
// scheduler needed. See ../bil/STRATEGY.md's "Emulator architecture"
// note and the plan this spike validates.
//
// Extended: also proves a SECOND sequential blocking async call works
// after the first, since Phase A's ripple demo does many in a row.
//
//go:build js && wasm

package main

import (
	"fmt"
	"sync"
	"syscall/js"
	"time"
)

// waitForSignal blocks the calling goroutine until the JS side calls
// back, without blocking anything else in this program. The JS
// function "waitForChange" is expected to return immediately (kicking
// off an Atomics.waitAsync chain) and invoke cb exactly once, later,
// when the watched memory word changes.
func waitForSignal(which int) {
	done := make(chan struct{})
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) any {
		cb.Release()
		close(done)
		return nil
	})
	js.Global().Call("waitForChange", which, cb)
	<-done
}

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		fmt.Println("spike: goroutine A calling waitForSignal(0)")
		waitForSignal(0)
		fmt.Println("spike: goroutine A woke up from signal 0")
		fmt.Println("spike: goroutine A calling waitForSignal(1) -- SECOND sequential wait")
		waitForSignal(1)
		fmt.Println("spike: goroutine A woke up from signal 1 -- second wait also resolved")
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 12; i++ {
			fmt.Println("spike: heartbeat", i)
			time.Sleep(400 * time.Millisecond)
		}
	}()

	wg.Wait()
	fmt.Println("spike: done -- both goroutines finished cleanly, no deadlock")
}
