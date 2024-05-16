package store

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/kavinaravind/go-raft-message-queue/ds"
)

type MockSnapshotSink struct {
	buffer bytes.Buffer
	closed bool
}

func (m *MockSnapshotSink) ID() string {
	return "mock"
}

func (m *MockSnapshotSink) Cancel() error {
	m.closed = true
	return nil
}

func (m *MockSnapshotSink) Close() error {
	m.closed = true
	return nil
}

func (m *MockSnapshotSink) Write(p []byte) (n int, err error) {
	return m.buffer.Write(p)
}

func queuesAreEqual(q1, q2 *ds.Queue[int]) bool {
	if len(q1.Messages) != len(q2.Messages) {
		return false
	}

	for i := 0; i < len(q1.Messages); i++ {
		if q1.Messages[i] != q2.Messages[i] {
			return false
		}
	}

	return true
}

func TestPersist(t *testing.T) {
	queue := ds.NewQueue[int]()

	queue.Enqueue(ds.Message[int]{Data: 1})
	queue.Enqueue(ds.Message[int]{Data: 2})
	queue.Enqueue(ds.Message[int]{Data: 3})

	snapshot := Snapshot[int]{queue: queue}

	sink := &MockSnapshotSink{}

	err := snapshot.Persist(sink)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !sink.closed {
		t.Errorf("expected sink to be closed")
	}

	dec := gob.NewDecoder(&sink.buffer)
	var decodedQueue ds.Queue[int]
	if err := dec.Decode(&decodedQueue); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !queuesAreEqual(queue, &decodedQueue) {
		t.Errorf("expected %v, got %v", queue, &decodedQueue)
	}
}

func TestRelease(t *testing.T) {
	snapshot := Snapshot[int]{queue: ds.NewQueue[int]()}
	snapshot.Release() // should not panic
}
