package ds

import "sync"

type Message[T any] struct {
	Data T
}

type Queue[T any] struct {
	messages []Message[T]
	lock     sync.RWMutex
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{}
}

func (q *Queue[T]) Enqueue(message Message[T]) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.messages = append(q.messages, message)
}

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
