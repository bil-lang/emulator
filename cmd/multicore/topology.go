package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Same fixed link-index convention as cmd/wasm/static/index.html's buildTopology:
// a node's link[i] slots are 0=north, 1=east, 2=south, 3=west. Plain mesh,
// not torus -- edge/corner nodes simply have fewer wired slots.
const (
	north        = 0
	east         = 1
	south        = 2
	west         = 3
	numLinkSlots = 4
)

// nodeEnv accumulates the extra BIL_LINK_*_OUT/_IN env entries and the
// set of wired indices for one grid node.
type nodeEnv struct {
	extra []string
	wired map[int]bool
}

// buildTopology allocates one Unix-domain-socket path per directed edge
// (the native analog of cmd/wasm/static/index.html's makeLinkPair() ab/ba
// SharedArrayBuffer pair) under sockDir, and returns each node's extra
// env vars for cmd/multicore to pass to its spawned process.
func buildTopology(rows, cols int, sockDir string) [][]nodeEnv {
	grid := make([][]nodeEnv, rows)
	for r := range grid {
		grid[r] = make([]nodeEnv, cols)
		for c := range grid[r] {
			grid[r][c] = nodeEnv{wired: map[int]bool{}}
		}
	}

	// wire connects a directed data flow from (r1,c1)'s link[idx1] (its
	// Send/OUT side) to (r2,c2)'s link[idx2] (its Recv/IN side) over one
	// dedicated socket. Whichever of the two nodes has the smaller flat
	// ID listens (binds first); the other dials with retry/backoff --
	// see bilink/bilink_native.go. This choice is independent of
	// which side sends/receives on this particular socket, so it applies
	// uniformly to both directed sockets of one edge.
	sockN := 0
	wire := func(r1, c1, idx1, r2, c2, idx2 int) {
		sockN++
		path := filepath.Join(sockDir, fmt.Sprintf("link-%d.sock", sockN))
		id1, id2 := r1*cols+c1, r2*cols+c2
		listenerIsSrc := id1 < id2
		srcMode, dstMode := "dial", "dial"
		if listenerIsSrc {
			srcMode = "listen"
		} else {
			dstMode = "listen"
		}
		grid[r1][c1].extra = append(grid[r1][c1].extra,
			fmt.Sprintf("BIL_LINK_%d_OUT=%s:%s", idx1, srcMode, path))
		grid[r1][c1].wired[idx1] = true
		grid[r2][c2].extra = append(grid[r2][c2].extra,
			fmt.Sprintf("BIL_LINK_%d_IN=%s:%s", idx2, dstMode, path))
		grid[r2][c2].wired[idx2] = true
	}

	for r := 0; r < rows; r++ {
		for c := 0; c < cols-1; c++ {
			wire(r, c, east, r, c+1, west) // eastbound data
			wire(r, c+1, west, r, c, east) // westbound data
		}
	}
	for r := 0; r < rows-1; r++ {
		for c := 0; c < cols; c++ {
			wire(r, c, south, r+1, c, north) // southbound data
			wire(r+1, c, north, r, c, south) // northbound data
		}
	}
	return grid
}

// wiredList renders a nodeEnv's wired indices as BIL_WIRED's expected
// comma-separated form.
func (n nodeEnv) wiredList() string {
	s := ""
	for idx := 0; idx < numLinkSlots; idx++ {
		if !n.wired[idx] {
			continue
		}
		if s != "" {
			s += ","
		}
		s += fmt.Sprint(idx)
	}
	return s
}

func mustMkdirTemp(pattern string) string {
	dir, err := os.MkdirTemp("", pattern)
	if err != nil {
		fatalf("creating temp dir: %v", err)
	}
	return dir
}
