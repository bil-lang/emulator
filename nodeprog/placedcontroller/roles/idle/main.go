//go:build js && wasm

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:40
package main

import "sync"

import "emulator/nodeprog/bilink"

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:41

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:42
import "time"

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:43

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:44
const rounds = 3

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:45
const hopDelay = 150 * time.Millisecond

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:46

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:47
// controller runs at exactly one processor, (0,0): originates each

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:48
// round's value east on toEast, then waits for it to come back

// reflected on fromEast.
//
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:49
func
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:50
controller(toEast chan<- int32, fromEast <-chan int32) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:51
	cols := bilink.NumCols()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:52
	bilink.Screenf("controller ready, %d cols", cols)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:53
	for round := 0; round < rounds; round++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:54
		time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:55
		bilink.Send(1, int32(round))
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:55

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:56
		var v int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:57
		v = bilink.Recv(1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:57

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:58
		bilink.Screenf("round %d: reflected %d", round, v)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:59
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:60
	bilink.Screenf("controller done")

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:61
	time.Sleep(1<<63 - 1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:61

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:62
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:63

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:64
// rowEnd runs at (0,cols-1): the far end of row 0, receiving from the

// west and reflecting straight back west.
//
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:65
func
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:66
rowEnd(fromWest <-chan int32, toWest chan<- int32) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:67
	bilink.Screenf("row-end ready")

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:68
	for round := 0; round < rounds; round++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:69
		var v int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:70
		v = bilink.Recv(3)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:70

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:71
		bilink.Screenf("round %d: got %d, reflecting", round, v)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:72
		time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:73
		bilink.Send(3, v)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:73

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:74
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:75
	bilink.Screenf("row-end done")

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:76
	time.Sleep(1<<63 - 1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:76

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:77
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:78

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:79
// relay runs on every other row-0 processor: forward east, forward

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:80
// the reflection back west. Both the west and east links are used

// bidirectionally, so each gets two directional parameters.
//
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:81
func
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:82
relay(fromWestIn <-chan int32, toEastOut chan<- int32, toEastIn <-chan int32, fromWestOut chan<- int32) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:83
	bilink.Screenf("relay ready")

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:84
	for round := 0; round < rounds; round++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:85
		var v int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:86
		v = bilink.Recv(3)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:86

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:87
		bilink.Send(1, v)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:87

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:88
		var v2 int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:89
		v2 = bilink.Recv(1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:89

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:90
		bilink.Send(3, v2)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:90

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:91
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:92
	bilink.Screenf("relay done")

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:93
	time.Sleep(1<<63 - 1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:93

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:94
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:95

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:96
// idle runs everywhere off row 0 — nothing to do in this demo, just

// confirming placement dispatched here correctly at all.
//
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:97
func
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:98
idle() {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:99
	bilink.Screenf("idle -- not part of row 0")

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:100
	time.Sleep(1<<63 - 1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:100

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:101
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:102

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:103
func main() {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:104
	r, cols := bilink.Row(), bilink.NumCols()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:105

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:106
	_ = (0)
	_ = (0)
	_ = (0)
	_ = (cols - 1)
	_ = (r == 0)
	_ = (r == 0)
	idle()
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:128

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/20-placed-controller.bil:129
}

func par(branches ...func()) {
	var wg sync.WaitGroup
	wg.Add(len(branches))
	for _, b := range branches {
		go func(b func()) {
			defer wg.Done()
			b()
		}(b)
	}
	wg.Wait()
}

func parFor(n int, body func(int)) {
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			body(i)
		}(i)
	}
	wg.Wait()
}

func makeChans[T any](n int) []chan T {
	cs := make([]chan T, n)
	for i := range cs {
		cs[i] = make(chan T)
	}
	return cs
}

func splitN[T any](s []T, n int) [][]T {
	chunk := len(s) / n
	out := make([][]T, n)
	for i := range n {
		lo := i * chunk
		hi := lo + chunk
		if i == n-1 {
			hi = len(s)
		}
		out[i] = s[lo:hi:hi]
	}
	return out
}
