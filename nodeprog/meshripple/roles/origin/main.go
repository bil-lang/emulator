//go:build js && wasm

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:37
package main

import "sync"

import "emulator/nodeprog/bilink"

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:38

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:39
import "time"

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:40

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:41
const waves = 3

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:42

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:43
// hopDelay is purely cosmetic -- the rendezvous itself already

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:44
// enforces correct pacing (both sides must be present), this just

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:45
// makes the ripple visible to a human watching the grid rather than

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:46
// resolving instantly.

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:47
const hopDelay = 150 * time.Millisecond

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:48

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:49
// origin runs at column 0 of every row: originates each wave east,

// then awaits its reflection back on the same physical link.
//
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:50
func
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:51
origin(eastOut chan<- int32, eastIn <-chan int32) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:52
	r := bilink.Row()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:53
	bilink.Screenf("(%d,0) ready", r)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:54

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:55
	if bilink.NumCols() < 2 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:56
		bilink.Screenf("(%d,0) single column, nothing to ripple", r)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:57
		time.Sleep(1<<63 - 1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:57

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:58
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:59

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:60
	for wave := range waves {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:61
		time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:62
		bilink.Screenf("wave %d: send east = %d", wave, wave)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:63
		bilink.Send(1, int32(wave))
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:63

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:64
		var v int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:65
		v = bilink.Recv(1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:65

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:66
		bilink.Screenf("wave %d: reflected back = %d", wave, v)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:67
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:68

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:69
	bilink.Screenf("(%d,0) done, %d waves complete", r, waves)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:70
	time.Sleep(1<<63 - 1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:70

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:71
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:72

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:73
// reflect runs at the last column of every row: receives each wave

// from the west, then reflects it straight back.
//
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:74
func
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:75
reflect(westIn <-chan int32, westOut chan<- int32) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:76
	r, c := bilink.Row(), bilink.Col()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:77
	bilink.Screenf("(%d,%d) ready", r, c)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:78

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:79
	for wave := range waves {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:80
		var v int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:81
		v = bilink.Recv(3)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:81

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:82
		bilink.Screenf("wave %d: got %d, reflecting", wave, v)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:83
		time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:84
		bilink.Send(3, v)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:84

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:85
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:86

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:87
	bilink.Screenf("(%d,%d) done, %d waves complete", r, c, waves)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:88
	time.Sleep(1<<63 - 1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:88

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:89
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:90

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:91
// relay runs everywhere else: forwards each wave east, then forwards

// its reflection back west.
//
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:92
func
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:93
relay(westIn <-chan int32, eastOut chan<- int32, eastIn <-chan int32, westOut chan<- int32) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:94
	r, c := bilink.Row(), bilink.Col()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:95
	bilink.Screenf("(%d,%d) ready", r, c)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:96

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:97
	for wave := range waves {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:98
		var v int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:99
		v = bilink.Recv(3)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:99

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:100
		bilink.Screenf("wave %d: got %d, forward east", wave, v)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:101
		time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:102
		bilink.Send(1, v)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:102

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:103

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:104
		var v2 int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:105
		v2 = bilink.Recv(1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:105

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:106
		bilink.Screenf("wave %d: reflection %d, forward west", wave, v2)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:107
		time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:108
		bilink.Send(3, v2)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:108

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:109
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:110

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:111
	bilink.Screenf("(%d,%d) done, %d waves complete", r, c, waves)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:112
	time.Sleep(1<<63 - 1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:112

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:113
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:114

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:115
func main() {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:116
	cols := bilink.NumCols()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:117

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:118
	_ = (0)
	_ = (cols - 1)
	origin(nil, nil)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:136

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/19-mesh-ripple.bil:137
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
