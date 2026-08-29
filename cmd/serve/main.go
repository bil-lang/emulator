// serve is a minimal static file server that sets the two response
// headers SharedArrayBuffer requires (COOP/COEP) -- without them,
// SharedArrayBuffer does not exist in the page at all. Stdlib only,
// matching bil/tools/bilc's zero-third-party-dependency convention.
package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", "localhost:8787", "listen address")
	dir := flag.String("dir", ".", "directory to serve")
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
