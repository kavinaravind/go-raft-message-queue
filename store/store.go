package store

import (
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

// implement the raft fsm
func (s *Store[T]) Apply(log *raft.Log) interface{} {
	return nil
}

type Snapshot[T any] struct {
	queue *ds.Queue[T]
}

func (s *Snapshot[T]) Persist(sink raft.SnapshotSink) error {
	// Implement the Persist method
	return nil
}

func (s *Snapshot[T]) Release() {
	// Implement the Release method
}

func (s *Store[T]) Snapshot() (raft.FSMSnapshot, error) {
	// Copy the queue
	queue := s.queue.Copy()

	return &Snapshot[T]{queue: queue}, nil
}

func (s *Store[T]) Restore(rc io.ReadCloser) error {
	queue := ds.NewQueue[T]()

	dec := json.NewDecoder(rc)

	// Read and enqueue items until the JSON data is exhausted
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
