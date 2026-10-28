package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/child"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/front"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/adaptermap"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/startup"
)

func main() {
	log.SetOutput(os.Stderr)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	exeDir := filepath.Dir(exe)

	dataDir := os.Getenv("ENGRAM_DATA_DIR")
	project := os.Getenv("ENGRAM_PROJECT")
	if err := child.ValidateStartup(dataDir, project); err != nil {
		log.Fatal(err)
	}

	pinPath := startup.PinPath(exe)
	engramBin, err := startup.ResolveEngramBinary(exeDir, pinPath)
	if err != nil {
		log.Fatal(err)
	}

	sentinel := filepath.Join(dataDir, ".ce-sentinel-home")
	if err := child.MkdirSentinel(sentinel); err != nil {
		log.Fatal(err)
	}
	env, err := child.ChildEnv(child.Config{
		DataDir:      dataDir,
		Project:      project,
		SentinelHome: sentinel,
		EngramBin:    engramBin,
	})
	if err != nil {
		log.Fatal(err)
	}

	client, err := child.StartClient(ctx, engramBin, project, env, sentinel)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	svc := &adaptermap.Service{Client: client}
	if err := front.RunHost(ctx, svc); err != nil {
		log.Fatal(err)
	}
}
