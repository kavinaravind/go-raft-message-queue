package store

import (
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
func (s *Store[data]) Apply(log *raft.Log) interface{} {
	return nil
}

func (s *Store[data]) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}

func (s *Store[data]) Restore(io.ReadCloser) error {
	return nil
}
