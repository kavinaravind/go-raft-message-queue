package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kavinaravind/go-raft-message-queue/consensus"
	"github.com/kavinaravind/go-raft-message-queue/model"
	"github.com/kavinaravind/go-raft-message-queue/store"
)

func setup(t *testing.T) (*store.Store[model.Comment], *Server) {
	// Create a new logger
	logger := slog.Default()

	// Create a new store
	store := store.NewStore[model.Comment](logger)

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
		<-shutdownComplete
	})

	// Create a new server with a mock store and logger
	server := NewServer(store, logger)

	return store, server
}

func TestServer(t *testing.T) {
	store, server := setup(t)

	if err := store.WaitForNodeToBeLeader(5 * time.Second); err != nil {
		t.Fatalf("expected node1 to be leader, got: %v", err)
	}

	t.Run("Initialize", func(t *testing.T) {
		// Create a context with a cancel function
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel() // Ensure cancel is called to clean up resources

		// Initialize the server
		shutdownServerComplete := server.Initialize(ctx, &Config{Address: ":8080"})

		// Register a cleanup function to shut down the server
		t.Cleanup(func() {
			// Cancel the context to trigger server shutdown
			cancel()

			// Wait for shutdown to complete
			<-shutdownServerComplete
		})

	})

	t.Run("HandleSend", func(t *testing.T) {
		// Create a new HTTP request
		req, err := http.NewRequest(http.MethodPost, "/send", strings.NewReader(`{"id": "1", "text": "test"}`))
		if err != nil {
			t.Fatal(err)
		}

		// Create a ResponseRecorder to record the response
		rr := httptest.NewRecorder()

		// Call the handler function
		handler := http.HandlerFunc(server.handleSend)
		handler.ServeHTTP(rr, req)

		// Check the status code is what we expect
		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
		}
	})

	t.Run("HandleRecieve", func(t *testing.T) {
		// Create a new HTTP request
		req, err := http.NewRequest(http.MethodGet, "/recieve", nil)
		if err != nil {
			t.Fatal(err)
		}

		// Create a ResponseRecorder to record the response
		rr := httptest.NewRecorder()

		// Call the handler function
		handler := http.HandlerFunc(server.handleRecieve)
		handler.ServeHTTP(rr, req)

		// Check the status code is what we expect
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
	})

	t.Run("HandleStats", func(t *testing.T) {
		// Create a new HTTP request
		req, err := http.NewRequest(http.MethodGet, "/stats", nil)
		if err != nil {
			t.Fatal(err)
		}

		// Create a ResponseRecorder to record the response
		rr := httptest.NewRecorder()

		// Call the handler function
		handler := http.HandlerFunc(server.handleStats)
		handler.ServeHTTP(rr, req)

		// Check the status code is what we expect
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Check state
		var stats map[string]string
		if err := json.NewDecoder(rr.Body).Decode(&stats); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if stats["state"] != "Leader" {
			t.Errorf("expected node2 to be a follower, got: %v", stats["state"])
		}
	})

	t.Run("HandleJoin", func(t *testing.T) {
		// Create a new HTTP request
		req, err := http.NewRequest(http.MethodPost, "/join", strings.NewReader(`{"address": "localhost:8001", "id": "node2"}`))
		if err != nil {
			t.Fatal(err)
		}

		// Create a ResponseRecorder to record the response
		rr := httptest.NewRecorder()

		// Call the handler function
		handler := http.HandlerFunc(server.handleJoin)
		handler.ServeHTTP(rr, req)

		// Check the status code is what we expect (will be 201 but raft will to appendEntries as node2 is not up)
		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
		}
	})
}
