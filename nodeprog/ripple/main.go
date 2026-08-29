// ripple is the Phase A substrate proof: a counter travels east along
// each row, reflects at the last column, and travels back west. Every
// node uses exactly one link direction at a time, so this exercises
// the real blocking rendezvous end-to-end without also needing
// multi-way alt (see ../../README.md).
//
//go:build js && wasm

package main

import (
	"time"

	"emulator/nodeprog/bilink"
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
			// originate the wave east, then await its reflection
			time.Sleep(hopDelay)
			bilink.Screenf("wave %d: send east = %d", wave, wave)
			bilink.Send(bilink.East, int32(wave))
			v := bilink.Recv(bilink.East)
			bilink.Screenf("wave %d: reflected back = %d", wave, v)

		case c == cols-1:
			// last column: receive, then reflect straight back west
			v := bilink.Recv(bilink.West)
			bilink.Screenf("wave %d: got %d, reflecting", wave, v)
			time.Sleep(hopDelay)
			bilink.Send(bilink.West, v)

		default:
			// middle: forward the wave east, then forward its reflection west
			v := bilink.Recv(bilink.West)
			bilink.Screenf("wave %d: got %d, forward east", wave, v)
			time.Sleep(hopDelay)
			bilink.Send(bilink.East, v)

			v2 := bilink.Recv(bilink.East)
			bilink.Screenf("wave %d: reflection %d, forward west", wave, v2)
			time.Sleep(hopDelay)
			bilink.Send(bilink.West, v2)
		}
	}

	bilink.Screenf("(%d,%d) done, %d waves complete", r, c, waves)
	select {} // stay resident so the final screen state remains visible
}
