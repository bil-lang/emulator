// multicore is the native counterpart to cmd/wasm: instead of one browser
// Worker per grid node, it runs one native OS process per grid node, each
// hinted onto its own CPU core (a real pin on Linux, best-effort/no-op
// elsewhere -- see ../../bilink/affinity_linux.go and README.md).
//
// Unlike `bil emu` (the sibling bil repo's launcher), multicore expects a
// nodeprog directory already built by bilc -- either a single main.go, or
// roles/*/main.go + roles/placement.json -- exactly the on-disk shape
// ../../README.md's Quickstart already produces. It does not invoke bilc
// itself. See ./README.md for this backend's own quirks and quickstart.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := newFlagSet()
	dir, rows, cols, err := fs.parse(args)
	if err != nil {
		return 2
	}

	numCPU := runtime.NumCPU()
	if rows*cols > numCPU {
		fmt.Fprintf(os.Stderr, "multicore: %dx%d grid needs %d processes, but this machine has only %d CPU cores\n", rows, cols, rows*cols, numCPU)
		return 1
	}

	buildDir := mustMkdirTemp("multicore-build-*")
	defer os.RemoveAll(buildDir)
	sockDir := mustMkdirTemp("multicore-sock-*")
	defer os.RemoveAll(sockDir)

	roleBinary, roleFor, err := prepare(dir, buildDir, rows, cols)
	if err != nil {
		fmt.Fprintln(os.Stderr, "multicore:", err)
		return 1
	}

	topo := buildTopology(rows, cols, sockDir)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var stdoutMu sync.Mutex
	out := &prefixWriter{dst: os.Stdout, mu: &stdoutMu}
	errOut := &prefixWriter{dst: os.Stderr, mu: &stdoutMu}

	type child struct {
		r, c int
		cmd  *exec.Cmd
	}
	var children []child
	coreHint := 0
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			role, err := roleFor(r, c)
			if err != nil {
				fmt.Fprintln(os.Stderr, "multicore:", err)
				return 1
			}
			bin, ok := roleBinary[role]
			if !ok {
				fmt.Fprintf(os.Stderr, "multicore: node (%d,%d) resolved to role %q, but no such binary was built\n", r, c, role)
				return 1
			}

			cmd := exec.CommandContext(ctx, bin)
			ne := topo[r][c]
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("BIL_ROW=%d", r),
				fmt.Sprintf("BIL_COL=%d", c),
				fmt.Sprintf("BIL_ROWS=%d", rows),
				fmt.Sprintf("BIL_COLS=%d", cols),
				fmt.Sprintf("BIL_NUM_LINKS=%d", numLinkSlots),
				fmt.Sprintf("BIL_WIRED=%s", ne.wiredList()),
				fmt.Sprintf("BIL_CORE_HINT=%d", coreHint),
			)
			cmd.Env = append(cmd.Env, ne.extra...)
			cmd.Stdout = &nodeWriter{w: out, r: r, c: c}
			cmd.Stderr = &nodeWriter{w: errOut, r: r, c: c}

			if err := cmd.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "multicore: starting node (%d,%d): %v\n", r, c, err)
				return 1
			}
			children = append(children, child{r, c, cmd})
			coreHint++
		}
	}

	fmt.Fprintf(os.Stdout, "multicore: running %dx%d grid (%d processes, %d CPU cores available) -- Ctrl-C to stop\n", rows, cols, rows*cols, numCPU)

	failed := false
	for _, ch := range children {
		if err := ch.cmd.Wait(); err != nil {
			if ctx.Err() != nil {
				continue // stopped via Ctrl-C/SIGTERM, not a real failure
			}
			fmt.Fprintf(os.Stderr, "multicore: node (%d,%d) exited: %v\n", ch.r, ch.c, err)
			failed = true
		}
	}
	if failed {
		return 1
	}
	return 0
}

// prepare builds dir's node program(s) natively into buildDir, returning
// each role's compiled binary path and a function that resolves which
// role runs at a given grid position. A dir with roles/placement.json builds
// one binary per role subdirectory; a plain dir (a single main.go, no
// placement) builds one binary that runs at every position.
func prepare(dir, buildDir string, rows, cols int) (map[string]string, func(r, c int) (string, error), error) {
	placementPath := filepath.Join(dir, "roles", "placement.json")
	if _, err := os.Stat(placementPath); err == nil {
		m, err := loadManifest(placementPath)
		if err != nil {
			return nil, nil, err
		}
		entries, err := os.ReadDir(filepath.Join(dir, "roles"))
		if err != nil {
			return nil, nil, err
		}
		bins := map[string]string{}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			role := e.Name()
			out := filepath.Join(buildDir, role)
			if err := goBuildNative(filepath.Join(dir, "roles", role), out); err != nil {
				return nil, nil, fmt.Errorf("building role %q: %w", role, err)
			}
			bins[role] = out
		}
		roleFor := func(r, c int) (string, error) { return resolveRole(m, r, c, rows, cols) }
		return bins, roleFor, nil
	}

	if _, err := os.Stat(filepath.Join(dir, "main.go")); err != nil {
		return nil, nil, fmt.Errorf("%s has neither roles/placement.json nor main.go -- not a built nodeprog dir", dir)
	}
	out := filepath.Join(buildDir, "node")
	if err := goBuildNative(dir, out); err != nil {
		return nil, nil, fmt.Errorf("building %s: %w", dir, err)
	}
	return map[string]string{"node": out}, func(r, c int) (string, error) { return "node", nil }, nil
}

// goBuildNative builds the Go package in srcDir as a plain native binary
// (no GOOS/GOARCH override -- the whole point of this backend), matching
// bil emu's own goBuildWasm helper in shape.
func goBuildNative(srcDir, out string) error {
	cmd := exec.Command("go", "build", "-o", out, ".")
	cmd.Dir = srcDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w\n%s", err, output)
	}
	return nil
}

// prefixWriter serializes writes from multiple concurrent sources (one
// per child process) onto one shared destination, so lines from
// different nodes never interleave mid-line.
type prefixWriter struct {
	dst io.Writer
	mu  *sync.Mutex
}

// nodeWriter prefixes every line with its node's (row,col) before handing
// it to the shared prefixWriter. bilink_native.go's Screenf already
// prefixes its own output this way, but a node's raw stderr (a Go panic,
// a build/runtime error) doesn't, so this covers that case too.
type nodeWriter struct {
	w    *prefixWriter
	r, c int
}

func (n *nodeWriter) Write(p []byte) (int, error) {
	n.w.mu.Lock()
	defer n.w.mu.Unlock()
	start := 0
	for i, b := range p {
		if b == '\n' {
			line := p[start:i]
			if len(line) == 0 || line[0] != '[' {
				fmt.Fprintf(n.w.dst, "[%d,%d] %s\n", n.r, n.c, line)
			} else {
				n.w.dst.Write(p[start : i+1])
			}
			start = i + 1
		}
	}
	if start < len(p) {
		fmt.Fprintf(n.w.dst, "[%d,%d] %s\n", n.r, n.c, p[start:])
	}
	return len(p), nil
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "multicore: "+format+"\n", a...)
	os.Exit(1)
}
