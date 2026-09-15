// Native counterpart to bilink.go: the same link/channel primitive, but
// backed by real OS-level IPC between native processes instead of
// syscall/js + SharedArrayBuffer/Atomics -- see ../cmd/multicore, the
// launcher that sets every env var this file reads and spawns one native
// OS process per grid node (one per real CPU core, on a machine that has
// enough of them -- see ../cmd/multicore/README.md).
//
// Each wired link[idx] is backed by two Unix-domain-socket connections,
// one per direction -- the direct analog of cmd/wasm/static/index.html's
// makeLinkPair() ab/ba SharedArrayBuffer pair. Send blocks writing a
// 4-byte payload then reading a 1-byte ack; Recv blocks reading 4 bytes,
// writes a 1-byte ack, and returns immediately -- matching
// node-worker.js's linkSend/linkRecv handshake exactly (sender parks
// until the receiver has consumed; receiver returns as soon as it has
// read+acked, without waiting for the sender to reset). Because each
// direction owns a dedicated connection and the protocol never has more
// than one payload outstanding per connection, no message framing is
// needed -- unlike a single bidirectional connection shared by both
// directions, where a receiver's ack bytes and a sender's payload bytes
// would land on the same ordered stream with no way to tell them apart.
//
//go:build !js

package bilink

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	row, col, numRows, numCols, numLinks int
	wiredSet                             map[int]bool
	outConns                             []net.Conn
	inConns                              []net.Conn
)

func init() {
	row = mustEnvInt("BIL_ROW")
	col = mustEnvInt("BIL_COL")
	numRows = mustEnvInt("BIL_ROWS")
	numCols = mustEnvInt("BIL_COLS")
	numLinks = mustEnvInt("BIL_NUM_LINKS")

	wiredSet = map[int]bool{}
	if w := os.Getenv("BIL_WIRED"); w != "" {
		for _, s := range strings.Split(w, ",") {
			idx, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil {
				fatalf("bilink: bad BIL_WIRED entry %q: %v", s, err)
			}
			wiredSet[idx] = true
		}
	}

	outConns = make([]net.Conn, numLinks)
	inConns = make([]net.Conn, numLinks)

	done := make(chan struct{})
	pending := 0
	for idx := range wiredSet {
		if spec := os.Getenv(fmt.Sprintf("BIL_LINK_%d_OUT", idx)); spec != "" {
			pending++
			go func(idx int, spec string) {
				outConns[idx] = mustDialOrListen(idx, "OUT", spec)
				done <- struct{}{}
			}(idx, spec)
		}
		if spec := os.Getenv(fmt.Sprintf("BIL_LINK_%d_IN", idx)); spec != "" {
			pending++
			go func(idx int, spec string) {
				inConns[idx] = mustDialOrListen(idx, "IN", spec)
				done <- struct{}{}
			}(idx, spec)
		}
	}
	for i := 0; i < pending; i++ {
		<-done
	}

	// Lock this goroutine (the one that will go on to run main()) to its
	// current OS thread before applying the core-affinity hint below --
	// otherwise Go's scheduler is free to migrate it to a different OS
	// thread later, silently defeating the hint. See affinity_linux.go
	// (the only platform with a real hint) and affinity_stub.go.
	runtime.LockOSThread()
	applyCoreAffinityHint()
}

// mustDialOrListen connects one directional link socket. spec is
// "listen:<path>" (bind and accept once -- used by whichever side of an
// edge multicore designates the sender/listener) or "dial:<path>" (connect,
// retrying with a short backoff since sibling processes are spawned
// concurrently and the listener may not have bound yet).
func mustDialOrListen(idx int, dir, spec string) net.Conn {
	mode, path, ok := strings.Cut(spec, ":")
	if !ok {
		fatalf("bilink: malformed BIL_LINK_%d_%s %q (want \"listen:<path>\" or \"dial:<path>\")", idx, dir, spec)
	}
	switch mode {
	case "listen":
		os.Remove(path) // clear a stale socket file from a prior aborted run
		ln, err := net.Listen("unix", path)
		if err != nil {
			fatalf("bilink: link[%d] %s: listen %s: %v", idx, dir, path, err)
		}
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			fatalf("bilink: link[%d] %s: accept on %s: %v", idx, dir, path, err)
		}
		return conn
	case "dial":
		deadline := time.Now().Add(10 * time.Second)
		for {
			conn, err := net.Dial("unix", path)
			if err == nil {
				return conn
			}
			if time.Now().After(deadline) {
				fatalf("bilink: link[%d] %s: dial %s: timed out: %v", idx, dir, path, err)
			}
			time.Sleep(20 * time.Millisecond)
		}
	default:
		fatalf("bilink: link[%d] %s: unknown mode %q in %q", idx, dir, mode, spec)
		panic("unreachable")
	}
}

// Send blocks until v has been both delivered to, and acknowledged as
// consumed by, whatever this node is wired to via link[idx] -- see
// bilink.go's Send doc comment for the semantics this must match.
func Send(idx int, v int32) {
	c := outConns[idx]
	if c == nil {
		panic(fmt.Sprintf("bilink: Send on unwired link[%d]", idx))
	}
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(v))
	if _, err := c.Write(buf[:]); err != nil {
		fatalf("bilink: link[%d] out: write: %v", idx, err)
	}
	var ack [1]byte
	if _, err := io.ReadFull(c, ack[:]); err != nil {
		fatalf("bilink: link[%d] out: read ack: %v", idx, err)
	}
}

// Recv blocks until a value arrives via link[idx].
func Recv(idx int) int32 {
	c := inConns[idx]
	if c == nil {
		panic(fmt.Sprintf("bilink: Recv on unwired link[%d]", idx))
	}
	var buf [4]byte
	if _, err := io.ReadFull(c, buf[:]); err != nil {
		fatalf("bilink: link[%d] in: read: %v", idx, err)
	}
	v := int32(binary.LittleEndian.Uint32(buf[:]))
	if _, err := c.Write([]byte{1}); err != nil {
		fatalf("bilink: link[%d] in: write ack: %v", idx, err)
	}
	return v
}

func Row() int      { return row }
func Col() int      { return col }
func NumRows() int  { return numRows }
func NumCols() int  { return numCols }
func NumLinks() int { return numLinks }
func ID() int       { return row*numCols + col }

// LinkWired reports whether link[idx] actually has a neighbour wired up
// on this node -- see bilink.go's LinkWired doc comment.
func LinkWired(idx int) bool { return wiredSet[idx] }

// Screenf writes one line to stdout, prefixed with this node's position --
// multicore captures each child process's stdout and displays it, the
// native analog of node-worker.js's screenPrint/postMessage.
func Screenf(format string, a ...any) {
	fmt.Printf("[%d,%d] %s\n", row, col, fmt.Sprintf(format, a...))
}

func mustEnvInt(name string) int {
	s := os.Getenv(name)
	if s == "" {
		fatalf("bilink: required env var %s not set -- this binary must be launched by cmd/multicore, not run directly", name)
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		fatalf("bilink: env var %s=%q: %v", name, s, err)
	}
	return n
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}
