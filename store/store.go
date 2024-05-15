package store

import (
	"io"
	"log/slog"
	"time"

	"github.com/hashicorp/raft"
	"github.com/kavinaravind/go-raft-message-queue/ds"
)

const (
	Send = iota
	Recieve
)

type Comment struct {
	Timestamp time.Time `json:"timestamp,omitempty"`
	Author    string    `json:"author,omitempty"`
	Content   string    `json:"content,omitempty"`
}

type Store[T any] struct {
	queue *ds.Queue[T]

	raft *raft.Raft

	logger *slog.Logger
}

func NewStore[T any]() (*Store[T], error) {
	return &Store[T]{
		queue:  ds.NewQueue[T](),
		logger: slog.Default(),
	}, nil
}

func (s *Store[T]) Initialize(enableSingle bool, localID string) error {

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
