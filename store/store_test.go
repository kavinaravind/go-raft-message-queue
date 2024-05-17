package store

import (
	"bytes"
	"context"
	"encoding/gob"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/kavinaravind/go-raft-message-queue/consensus"
	"github.com/kavinaravind/go-raft-message-queue/ds"
	"github.com/kavinaravind/go-raft-message-queue/model"
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

func setup(t *testing.T) *Store[model.Comment] {
	// Create a new logger
	logger := slog.Default()

	// Create a new store
	store := NewStore[model.Comment](logger)

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

	// Create a context with a cancel function
	ctx, cancel := context.WithCancel(context.Background())

	// Call Initialize
	shutdownComplete, err := store.Initialize(ctx, conf)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	t.Cleanup(func() {
		// Cancel the context to trigger store shutdown
		cancel()

		// Wait for shutdown to complete
		select {
		case <-shutdownComplete:
			// Shutdown completed
		case <-time.After(5 * time.Second):
			t.Error("Timeout waiting for shutdown to complete")
		}
	})

	return store
}

func TestStore(t *testing.T) {
	store := setup(t)

	if err := store.WaitForNodeToBeLeader(5 * time.Second); err != nil {
		t.Fatalf("expected node1 to be leader, got: %v", err)
	}

	comment := model.Comment{
		Timestamp: nil,
		Author:    "Alice",
		Content:   "Hello, World!",
	}

	t.Run("Send", func(t *testing.T) {
		// Send a message
		if err := store.Send(comment); err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})

	t.Run("Recieve", func(t *testing.T) {
		// Recieve a message
		msg, err := store.Recieve()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if msg == nil {
			t.Fatal("Expected comment, got nil")
		}
		if msg.Data != comment {
			t.Errorf("Expected comment to be %v, got %v", comment, msg.Data)
		}
	})

	t.Run("Recieve (empty)", func(t *testing.T) {
		// Recieve a message
		msg, err := store.Recieve()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if (msg.Data != model.Comment{}) {
			t.Errorf("Expected nil, got %v", msg.Data)
		}
	})

	t.Run("Snapshot", func(t *testing.T) {
		// Push a message
		if err := store.Send(comment); err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Take a snapshot
		snapshot, err := store.Snapshot()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Check that the snapshot is correct
		if snapshot == nil {
			t.Fatal("Expected snapshot, got nil")
		}
	})

	t.Run("Restore", func(t *testing.T) {
		// Create a snapshot with some data
		queue := ds.NewQueue[model.Comment]()
		message := ds.Message[model.Comment]{Data: comment}
		queue.Enqueue(message)

		// Encode the snapshot
		buf := new(bytes.Buffer)
		enc := gob.NewEncoder(buf)
		err := enc.Encode(queue)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Restore from the snapshot
		if err := store.Restore(io.NopCloser(buf)); err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Check that the store's queue contains the data from the snapshot
		storeData, ok := store.queue.Dequeue()
		if !ok {
			t.Fatalf("Expected string, got: %v", ok)
		}

		if storeData != message {
			t.Errorf("Expected %v, got: %v", message, storeData.Data)
		}
	})
}
