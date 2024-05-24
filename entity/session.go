package entity

import (
	"fmt"
	"sync"
)

type Session[T comparable] struct {
	queue []T
	sync.Mutex
}

func NewSession[T comparable]() *Session[T] {
	return &Session[T]{queue: make([]T, 0)}
}

func (s *Session[T]) Add(t T) {
	s.Lock()
	defer s.Unlock()
	s.queue = append(s.queue, t)
}

func (s *Session[T]) Dequeue() (T, error) {
	s.Lock()
	defer s.Unlock()
	if len(s.queue) == 0 {
		var zeroT T
		return zeroT, fmt.Errorf("queue is empty")
	}
	top := s.queue[0]
	s.queue = s.queue[1:]
	return top, nil
}

func (s *Session[T]) CanMatch() bool {
	s.Lock()
	defer s.Unlock()
	return len(s.queue) >= 2
}
