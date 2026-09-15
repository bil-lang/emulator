package main

import (
	"flag"
	"fmt"
	"os"
)

type flagSet struct {
	fs *flag.FlagSet
}

func newFlagSet() *flagSet {
	fs := flag.NewFlagSet("multicore", flag.ContinueOnError)
	return &flagSet{fs: fs}
}

// parse reads -dir/-rows/-cols. Unlike `bil emu`'s -rows/-cols (0 = infer
// the smallest grid a placed-par program needs), multicore requires both
// explicit: this backend's grids are always small (capped at the
// machine's core count), so there's little to infer, and requiring it
// avoids duplicating bil emu's own inference logic in a second module.
func (f *flagSet) parse(args []string) (dir string, rows, cols int, err error) {
	dirFlag := f.fs.String("dir", "", "built nodeprog directory (a single main.go, or roles/*/main.go + roles/deploy.json)")
	rowsFlag := f.fs.Int("rows", 0, "grid rows")
	colsFlag := f.fs.Int("cols", 0, "grid cols")
	f.fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: multicore -dir <nodeprog dir> -rows N -cols N")
		f.fs.PrintDefaults()
	}
	if err := f.fs.Parse(args); err != nil {
		return "", 0, 0, err
	}
	if *dirFlag == "" || *rowsFlag <= 0 || *colsFlag <= 0 {
		f.fs.Usage()
		return "", 0, 0, fmt.Errorf("-dir, -rows, and -cols are all required")
	}
	return *dirFlag, *rowsFlag, *colsFlag, nil
}
