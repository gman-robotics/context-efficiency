package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/laya"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/shadow"
)

func main() {
	out := flag.String("out", "reports", "directory for the agreement report")
	flag.Parse()
	client, err := laya.NewClient(os.Getenv("LAYA_BASE_URL"), os.Getenv("LAYA_API_KEY"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "laya base url rejected")
		os.Exit(1)
	}
	rep, down, err := shadow.RunFixtures(context.Background(), client, time.Now())
	if errors.Is(err, laya.ErrKey) {
		fmt.Fprintln(os.Stderr, "laya api key not set")
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "laya shadow report failed")
		os.Exit(1)
	}
	if down {
		fmt.Fprintln(os.Stderr, "laya down")
		os.Exit(1)
	}
	if err := shadow.WriteReport(*out, rep); err != nil {
		fmt.Fprintln(os.Stderr, "laya shadow report failed")
		os.Exit(shadow.ReportExit(false, false))
	}
	os.Exit(shadow.ReportExit(true, rep.OverallPass))
}
