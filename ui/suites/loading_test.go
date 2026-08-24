package suites

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Hecatoncheir/lazyrest/finder"
)

func TestStartLoadCancelsPreviousLoad(t *testing.T) {
	widget := New()
	firstContext, firstID := widget.StartLoad()
	secondContext, secondID := widget.StartLoad()

	if !errors.Is(firstContext.Err(), context.Canceled) {
		t.Fatalf("previous load was not cancelled: %v", firstContext.Err())
	}
	if widget.IsCurrentLoad(firstID) || !widget.IsCurrentLoad(secondID) {
		t.Fatal("current load identifier was not updated")
	}

	widget.CancelLoad()
	if !errors.Is(secondContext.Err(), context.Canceled) {
		t.Fatalf("active load was not cancelled: %v", secondContext.Err())
	}
}

func TestLoadSuitesFromFileHonorsCancelledContext(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "request.http")
	if err := os.WriteFile(filePath, []byte("GET https://example.com\n"), 0644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result := New().LoadSuitesFromFile(ctx, finder.File{Name: "request.http", Path: filePath})
	if !errors.Is(result.Err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", result.Err)
	}
}

func TestFinishLoadRejectsStaleResult(t *testing.T) {
	widget := New()
	_, staleID := widget.StartLoad()
	_, currentID := widget.StartLoad()

	if widget.FinishLoad(staleID) {
		t.Fatal("stale load was accepted")
	}
	if !widget.FinishLoad(currentID) {
		t.Fatal("current load was rejected")
	}
}
