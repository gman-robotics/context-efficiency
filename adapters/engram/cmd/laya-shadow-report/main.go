package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/laya"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/shadow"
)

func main() {
	out := flag.String("out", "reports", "directory for the agreement report")
	flag.Parse()
	os.Exit(run(*out, os.Getenv("LAYA_BASE_URL"), os.Getenv("LAYA_API_KEY"), os.Stderr))
}

// run returns the process exit code. Messages never include the key.
func run(out, baseURL, apiKey string, stderr io.Writer) int {
	client, err := laya.NewClient(baseURL, apiKey)
	if err != nil {
		return fail(stderr, "laya base url rejected")
	}
	rep, down, err := shadow.RunFixtures(context.Background(), client, time.Now())
	switch {
	case errors.Is(err, laya.ErrKey):
		return fail(stderr, "laya api key not set")
	case err != nil:
		return fail(stderr, "laya shadow report failed")
	case down:
		return fail(stderr, "laya down")
	}
	if err := shadow.WriteReport(out, rep); err != nil {
		fmt.Fprintln(stderr, "laya shadow report failed")
		return shadow.ReportExit(false, false)
	}
	return shadow.ReportExit(true, rep.OverallPass)
}

func fail(stderr io.Writer, msg string) int {
	fmt.Fprintln(stderr, msg)
	return 1
}
