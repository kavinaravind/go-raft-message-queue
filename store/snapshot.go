package store

import (
	"encoding/gob"

	"github.com/hashicorp/raft"
	"github.com/kavinaravind/go-raft-message-queue/ds"
)

type Snapshot[T any] struct {
	queue *ds.Queue[T]
}

// Persist is used to persist the snapshot to the sink
func (s *Snapshot[T]) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// Create a gob encoder and encode the queue
		enc := gob.NewEncoder(sink)
		if err := enc.Encode(s.queue); err != nil {
			return err
		}

		return nil
	}()

	// If there was an error, cancel the sink and return the error
	if err != nil {
		sink.Cancel()
		return err
	}

	// Otherwise, close the sink and return any errors
	return sink.Close()
}

// Release is used to release any resources acquired during the snapshot
// In this case, we don't have any resources to clean up (noop)
func (s *Snapshot[T]) Release() {}
