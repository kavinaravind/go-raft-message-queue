package store

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/kavinaravind/go-raft-message-queue/consensus"
)

func TestNewStore(t *testing.T) {
	// Create a new logger
	logger := slog.Default()

	// Create a new store
	store := NewStore[int](logger)

	// Check that the store was created correctly
	if store.queue == nil {
		t.Error("Expected queue to be initialized, but it was nil")
	}
	if store.logger != logger {
		t.Error("Expected logger to be the same, but it was different")
	}
}

func TestStore_Initialize(t *testing.T) {
	// Create a new logger
	logger := slog.Default()

	// Create a new store
	store := NewStore[int](logger)

	tmpDir1, err := os.MkdirTemp("", "node1")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir1)
	})
	// Create a new consensus config
	conf := &consensus.Config{
		IsLeader:      true,
		ServerID:      "node1",
		BaseDirectory: tmpDir1,
		Address:       "localhost:8000",
	}

	// Create a context that will be cancelled after a delay
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Call Initialize
	shutdownComplete, err := store.Initialize(ctx, conf)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	t.Cleanup(func() {
		select {
		case <-shutdownComplete:
			// Shutdown completed
		case <-time.After(5 * time.Second):
			t.Error("Timeout waiting for shutdown to complete")
		}
	})

	// Check that the consensus field was set
	if store.consensus == nil {
		t.Error("Expected consensus to be set, but it was nil")
	}
}

// TODO: Add more tests
