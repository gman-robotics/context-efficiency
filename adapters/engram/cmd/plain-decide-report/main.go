package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/decide"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/shadow"
)

func main() {
	out := flag.String("out", "reports", "directory for the agreement report")
	flag.Parse()
	os.Exit(run(*out, os.Stderr))
}

// run exits 0 when every gated path passes, 2 when a gate fails, 1 on error.
func run(out string, stderr io.Writer) int {
	cases, err := shadow.LoadFixtures()
	if err != nil {
		fmt.Fprintln(stderr, "plain decide report: fixtures")
		return 1
	}
	rep, err := decide.Score(cases, time.Now())
	if err != nil {
		fmt.Fprintln(stderr, "plain decide report: score")
		return 1
	}
	if err := decide.Write(out, rep); err != nil {
		fmt.Fprintln(stderr, "plain decide report: write")
		return 1
	}
	if !rep.OverallPass {
		return 2
	}
	return 0
}
