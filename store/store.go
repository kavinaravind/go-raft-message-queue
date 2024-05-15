package store

import (
	"encoding/gob"
	"encoding/json"
	"io"
	"log/slog"

	"github.com/hashicorp/raft"
	"github.com/kavinaravind/go-raft-message-queue/consensus"
	"github.com/kavinaravind/go-raft-message-queue/ds"
)

const (
	Send = iota
	Recieve
)

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

func (s *Store[T]) Initialize(config *consensus.Config) error {
	s.logger.Info("Initializing store")

	raft, err := consensus.NewRaft(s, config)
	if err != nil {
		return err
	}

	s.raft = raft

	return nil
}

// implement the raft fsm interface

// Apply is used to apply a log entry to the store
func (s *Store[T]) Apply(log *raft.Log) interface{} {
	var message ds.Message[T]
	err := json.Unmarshal(log.Data, &message)
	if err != nil {
		s.logger.Error("failed to unmarshal message", "error", err)
		return false
	}

	switch message.Data.(type) {
	case Send:
		s.queue.Enqueue(message)
	case Recieve:
		s.queue.Dequeue()
	}

	return nil
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
