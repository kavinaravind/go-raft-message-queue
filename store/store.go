package store

import (
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

const (
	Send = iota
	Recieve
)

type command[T any] struct {
	Operation int           `json:"operation,omitempty"`
	Message   ds.Message[T] `json:"message,omitempty"`
}

func newCommand[T any](operation int, message ds.Message[T]) *command[T] {
	return &command[T]{
		Operation: operation,
		Message:   message,
	}
}

type Store[T any] struct {
	// ds that will be distributed across each node
	queue *ds.Queue[T]

	// raft instance that will be used to replicate the ds
	raft *raft.Raft

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
func (s *Store[T]) Initialize(config *consensus.Config) error {
	s.logger.Info("Initializing store")

	raft, err := consensus.NewRaft(s, config)
	if err != nil {
		return err
	}

	s.raft = raft

	return nil
}

// Send is used to enqueue a message into the queue
func (s *Store[T]) Send(data T) error {
	if s.raft.State() != raft.Leader {
		return errors.New("not the leader")
	}

	c := newCommand[T](Send, ds.Message[T]{Data: data})
	bytes, err := json.Marshal(c)
	if err != nil {
		return err
	}

	future := s.raft.Apply(bytes, 10*time.Second)
	return future.Error()
}

// Recieve is used to dequeue a message from the queue
func (s *Store[T]) Recieve() (*ds.Message[T], error) {
	if s.raft.State() != raft.Leader {
		return nil, errors.New("not the leader")
	}

	c := newCommand[T](Recieve, ds.Message[T]{})
	bytes, err := json.Marshal(c)
	if err != nil {
		s.logger.Error("failed to marshal message", "error", err)
		return nil, err
	}

	future := s.raft.Apply(bytes, 10*time.Second)
	if err := future.Error(); err != nil {
		s.logger.Error("failed to apply message", "error", err)
		return nil, err
	}

	switch response := future.Response().(type) {
	case error:
		// The Apply method returned an error
		return nil, response
	case *ds.Message[T]:
		// The Apply method returned a message
		return response, nil
	case nil:
		// The Apply method returned nil
		return nil, nil
	default:
		// The Apply method returned an unexpected type
		return nil, fmt.Errorf("unexpected response type: %T", response)
	}
}

// Stats is used to return the stats of the raft instance
func (s *Store[T]) Stats() map[string]string {
	return s.raft.Stats()
}

// Join is used to join a remote node to the raft cluster
func (s *Store[T]) Join(nodeID, address string) error {
	s.logger.Info("received join request for remote node %s at %s", nodeID, address)
	return consensus.Join(s.raft, nodeID, address)
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
		if !ok {
			s.logger.Error("failed to dequeue message")
			return fmt.Errorf("failed to dequeue message")
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
	queue := ds.NewQueue[T]()

	dec := gob.NewDecoder(rc)

	// Read and enqueue items until the data is exhausted
	var item ds.Message[T]
	for {
		if err := dec.Decode(&item); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		queue.Enqueue(item)
	}

	s.queue = queue
	return nil
}
