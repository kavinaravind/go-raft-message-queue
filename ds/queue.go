package ds

import "sync"

// Message is a generic message type
type Message[T any] struct {
	Data T
}

// Queue is a generic queue type
type Queue[T any] struct {
	Messages []Message[T]
	lock     sync.RWMutex
}

// NewQueue creates a new instance of the Queue
func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{}
}

// Enqueue is used to add a message to the queue
func (q *Queue[T]) Enqueue(message Message[T]) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.Messages = append(q.Messages, message)
}

// Dequeue is used to remove a message from the queue
func (q *Queue[T]) Dequeue() (Message[T], bool) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.Messages) == 0 {
		return Message[T]{}, false
	}

	message := q.Messages[0]
	q.Messages = q.Messages[1:]

	return message, true
}

// Copy is used to create a copy of the queue
func (q *Queue[T]) Copy() *Queue[T] {
	q.lock.RLock()
	defer q.lock.RUnlock()

	copy := NewQueue[T]()
	copy.Messages = append(copy.Messages, q.Messages...)

	return copy
}
