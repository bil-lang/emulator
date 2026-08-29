// ripple is the Phase A substrate proof: a counter travels east along
// each row, reflects at the last column, and travels back west. Every
// node uses exactly one link at a time, so this exercises the real
// blocking rendezvous end-to-end without also needing multi-way alt
// (see ../../README.md).
//
// bilink addresses links by plain index, not compass name -- north,
// east, south, west here are just local constants this program picks
// because that's the convention static/index.html's 2D-mesh topology
// happens to use (slot 0=north, 1=east, 2=south, 3=west). A different
// topology would define its own convention; bilink itself doesn't
// know or care what a given index physically connects to.
//
//go:build js && wasm

package main

import (
	"time"

	"emulator/nodeprog/bilink"
)

const (
	north = 0
	east  = 1
	south = 2
	west  = 3
)

const waves = 3

// hopDelay is purely cosmetic -- the rendezvous itself already
// enforces correct pacing (both sides must be present), this just
// makes the ripple visible to a human watching the grid rather than
// resolving instantly.
const hopDelay = 150 * time.Millisecond

func main() {
	r, c := bilink.Row(), bilink.Col()
	cols := bilink.NumCols()

	bilink.Screenf("(%d,%d) ready", r, c)

	if cols < 2 {
		bilink.Screenf("(%d,%d) single column, nothing to ripple", r, c)
		select {}
	}

	for wave := 0; wave < waves; wave++ {
		switch {
		case c == 0:
			// originate the wave east, then await its reflection back via the same link
			time.Sleep(hopDelay)
			bilink.Screenf("wave %d: send east = %d", wave, wave)
			bilink.Send(east, int32(wave))
			v := bilink.Recv(east)
			bilink.Screenf("wave %d: reflected back = %d", wave, v)

		case c == cols-1:
			// last column: receive, then reflect straight back west
			v := bilink.Recv(west)
			bilink.Screenf("wave %d: got %d, reflecting", wave, v)
			time.Sleep(hopDelay)
			bilink.Send(west, v)

		default:
			// middle: forward the wave east, then forward its reflection west
			v := bilink.Recv(west)
			bilink.Screenf("wave %d: got %d, forward east", wave, v)
			time.Sleep(hopDelay)
			bilink.Send(east, v)

			v2 := bilink.Recv(east)
			bilink.Screenf("wave %d: reflection %d, forward west", wave, v2)
			time.Sleep(hopDelay)
			bilink.Send(west, v2)
		}
	}

	bilink.Screenf("(%d,%d) done, %d waves complete", r, c, waves)
	select {} // stay resident so the final screen state remains visible
}
