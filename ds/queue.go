package ds

import "sync"

// Message is a generic message type
type Message[T any] struct {
	Data T
}

// Queue is a generic queue type
type Queue[T any] struct {
	messages []Message[T]
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

	q.messages = append(q.messages, message)
}

// Dequeue is used to remove a message from the queue
func (q *Queue[T]) Dequeue() (Message[T], bool) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.messages) == 0 {
		return Message[T]{}, false
	}

	message := q.messages[0]
	q.messages = q.messages[1:]

	return message, true
}

// Copy is used to create a copy of the queue
func (q *Queue[T]) Copy() *Queue[T] {
	q.lock.RLock()
	defer q.lock.RUnlock()

	copy := NewQueue[T]()
	copy.messages = append(copy.messages, q.messages...)

	return copy
}
