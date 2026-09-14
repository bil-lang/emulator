// wasm is the browser backend: a minimal static file server (its own
// static/ subdirectory, alongside this file) that sets the two response
// headers SharedArrayBuffer requires (COOP/COEP) -- without them,
// SharedArrayBuffer does not exist in the page at all. Stdlib only,
// matching bil/tools/bilc's zero-third-party-dependency convention. See
// ./README.md for this backend's own quickstart and quirks.
package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", "localhost:8787", "listen address")
	// Default assumes invocation from the repo root (`go run ./cmd/wasm`),
	// where this backend's own assets live right alongside this binary's
	// source -- pass -dir explicitly if running from anywhere else.
	dir := flag.String("dir", "cmd/wasm/static", "directory to serve")
	flag.Parse()

	fs := http.FileServer(http.Dir(*dir))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
		fs.ServeHTTP(w, r)
	})

	log.Printf("serving %s on http://%s (COOP/COEP set)", *dir, *addr)
	log.Fatal(http.ListenAndServe(*addr, handler))
}
