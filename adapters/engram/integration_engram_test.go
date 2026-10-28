//go:build integration

package engram_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/adaptermap"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/child"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/domain"
)

func TestRealEngramV360(t *testing.T) {
	bin := os.Getenv("ENGRAM_BIN")
	if bin == "" {
		t.Skip("ENGRAM_BIN not set")
	}
	dataDir := t.TempDir()
	sentinel := filepath.Join(dataDir, "sentinel-home")
	_ = os.MkdirAll(sentinel, 0o700)
	project := "ceci"
	env, err := child.ChildEnv(child.Config{
		DataDir:      dataDir,
		Project:      project,
		SentinelHome: sentinel,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	client, err := child.StartClient(ctx, bin, project, env, sentinel)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	svc := &adaptermap.Service{Client: client}

	items, err := svc.Recall(ctx, "anything", "bot", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("fresh store recall want empty, got %d", len(items))
	}
	h, err := svc.HandoffRead(ctx, "bot")
	if err != nil {
		t.Fatal(err)
	}
	if h != nil {
		t.Fatal("fresh handoff_read want null")
	}

	id, err := svc.Retain(ctx, "round-trip alpha", []string{"t"}, 2, "bot")
	if err != nil || id == "" {
		t.Fatalf("retain %v id %s", err, id)
	}
	items, err = svc.Recall(ctx, "alpha", "bot", 5)
	if err != nil || len(items) != 1 || items[0].Text != "round-trip alpha" {
		t.Fatalf("recall %+v err %v", items, err)
	}
	if items[0].CreatedAt == "" {
		t.Fatal("createdAt empty")
	}

	id2, err := svc.Retain(ctx, "second body differs", []string{"t"}, 1, "bot")
	id3, err2 := svc.Retain(ctx, "second body differs again", []string{"t"}, 1, "bot")
	if err != nil || err2 != nil || id2 == "" || id3 == "" || id2 == id3 {
		t.Fatalf("two retains same title clip should differ ids %s %s", id2, id3)
	}

	db := filepath.Join(dataDir, "engram.db")
	if _, err := os.Stat(db); err != nil {
		t.Fatalf("expected engram.db in data dir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(sentinel, "engram.db")); err == nil {
		t.Fatal("sentinel home must not hold engram.db")
	}

	_, err = svc.HandoffWrite(ctx, domain.Handoff{Goal: "g", OpenItems: []string{"<private>x</private>"}}, "bot")
	if err == nil {
		t.Fatal("handoff with private must error")
	}
}
