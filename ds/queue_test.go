package ds

import (
	"testing"
)

func TestNewQueue(t *testing.T) {
	q := NewQueue[int]()
	if len(q.messages) != 0 {
		t.Errorf("NewQueue() = %d; want 0", len(q.messages))
	}
}

func TestEnqueue(t *testing.T) {
	q := NewQueue[int]()
	q.Enqueue(Message[int]{Data: 1})

	if len(q.messages) != 1 || q.messages[0].Data != 1 {
		t.Errorf("Enqueue() = %v; want [1]", q.messages)
	}
}

func TestDequeue(t *testing.T) {
	q := NewQueue[int]()
	q.Enqueue(Message[int]{Data: 1})
	message, ok := q.Dequeue()

	if !ok || message.Data != 1 {
		t.Errorf("Dequeue() = %v, %v; want 1, true", message, ok)
	}
}

func TestCopy(t *testing.T) {
	q := NewQueue[int]()
	q.Enqueue(Message[int]{Data: 1})
	copy := q.Copy()

	q.lock.RLock()
	copy.lock.RLock()
	defer q.lock.RUnlock()
	defer copy.lock.RUnlock()

	if len(copy.messages) != 1 || copy.messages[0].Data != 1 {
		t.Errorf("Copy() = %v; want [1]", copy.messages)
	}
}
