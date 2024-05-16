package store

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/hashicorp/raft"
	"github.com/kavinaravind/go-raft-message-queue/consensus"
	"github.com/kavinaravind/go-raft-message-queue/ds"
)

// Specific operations that can be applied to the store
const (
	Send = iota
	Recieve
)

// command is used to represent the command that will be applied to the store
type command[T any] struct {
	Operation int           `json:"operation"`
	Message   ds.Message[T] `json:"message"`
}

// newCommand is used to create a new command instance
func newCommand[T any](operation int, message ds.Message[T]) *command[T] {
	return &command[T]{
		Operation: operation,
		Message:   message,
	}
}

type Store[T any] struct {
	// ds that will be distributed across each node
	queue *ds.Queue[T]

	// consensus instance that will be used to replicate the ds
	consensus *consensus.Consensus

	// logger instance
	logger *slog.Logger
}

// NewStore creates a new store instance with the given logger
func NewStore[T any](logger *slog.Logger) *Store[T] {
	return &Store[T]{
		queue:  ds.NewQueue[T](),
		logger: logger,
	}
}

// Initialize is used to initialize the store with the given config
func (s *Store[T]) Initialize(ctx context.Context, conf *consensus.Config) (chan struct{}, error) {
	s.logger.Info("Initializing store")

	consensus, err := consensus.NewConsensus(s, conf)
	if err != nil {
		return nil, err
	}

	s.consensus = consensus

	// Listen for context cancellation and shutdown the server
	shutdownComplete := make(chan struct{})
	go func() {
		<-ctx.Done()
		future := s.consensus.Node.Shutdown()
		if err := future.Error(); err != nil {
			s.logger.Error("Failed to shutdown node", "error", err)
		} else {
			s.logger.Info("Node shutdown")
		}
		close(shutdownComplete)
	}()

	return shutdownComplete, nil
}

// Send is used to enqueue a message into the queue
func (s *Store[T]) Send(data T) error {
	if s.consensus.Node.State() != raft.Leader {
		return errors.New("not the leader")
	}

	c := newCommand[T](Send, ds.Message[T]{Data: data})
	bytes, err := json.Marshal(c)
	if err != nil {
		return err
	}

	future := s.consensus.Node.Apply(bytes, 10*time.Second)
	return future.Error()
}

// Recieve is used to dequeue a message from the queue
func (s *Store[T]) Recieve() (*ds.Message[T], error) {
	if s.consensus.Node.State() != raft.Leader {
		return nil, errors.New("node is not the leader")
	}

	c := newCommand[T](Recieve, ds.Message[T]{})
	bytes, err := json.Marshal(c)
	if err != nil {
		s.logger.Error("failed to marshal message", "error", err)
		return nil, err
	}

	future := s.consensus.Node.Apply(bytes, 10*time.Second)
	if err := future.Error(); err != nil {
		s.logger.Error("failed to apply message", "error", err)
		return nil, err
	}

	switch response := future.Response().(type) {
	case nil:
		// The Apply method returned an empty response
		return &ds.Message[T]{}, nil
	case ds.Message[T]:
		// The Apply method returned a message
		return &response, nil
	default:
		// The Apply method returned an unexpected type
		return nil, fmt.Errorf("unexpected response type: %T", response)
	}
}

// Stats is used to return the stats of the raft instance
func (s *Store[T]) Stats() map[string]string {
	return s.consensus.Node.Stats()
}

// Join is used to join a remote node to the raft cluster
func (s *Store[T]) Join(nodeID, address string) error {
	s.logger.Info("received join request for remote node %s at %s", nodeID, address)
	return s.consensus.Join(nodeID, address)
}

// implement the raft fsm interface

// Apply is used to apply a log entry to the store
func (s *Store[T]) Apply(log *raft.Log) interface{} {
	var command command[T]
	err := json.Unmarshal(log.Data, &command)
	if err != nil {
		s.logger.Error("failed to unmarshal message", "error", err)
		return err
	}

	switch command.Operation {
	case Send:
		s.queue.Enqueue(command.Message)
		return nil
	case Recieve:
		val, ok := s.queue.Dequeue()
		// If the queue is empty, return nil
		if !ok {
			return nil
		}
		return val
	default:
		return fmt.Errorf("unknown operation: %v", command.Operation)
	}
}

// Snapshot is used to create a snapshot of the store
func (s *Store[T]) Snapshot() (raft.FSMSnapshot, error) {
	return &Snapshot[T]{
		queue: s.queue.Copy(),
	}, nil
}

// Restore is used to restore the store from a snapshot
func (s *Store[T]) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	dec := gob.NewDecoder(rc)

	// Decode the entire queue
	var queue ds.Queue[T]
	if err := dec.Decode(&queue); err != nil {
		return fmt.Errorf("failed to decode queue: %w", err)
	}

	s.queue = &queue

	return nil
}
