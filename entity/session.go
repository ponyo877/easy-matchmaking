package entity

import (
	"fmt"
	"sync"

	"golang.org/x/exp/maps"
)

type Session[T comparable] struct {
	queue map[T]bool
	sync.Mutex
}

func NewSession[T comparable]() *Session[T] {
	return &Session[T]{queue: make(map[T]bool)}
}

func (s *Session[T]) Add(t T) {
	s.Lock()
	defer s.Unlock()
	s.queue[t] = true
}

func (s *Session[T]) Dequeue() (T, error) {
	s.Lock()
	defer s.Unlock()
	if len(s.queue) == 0 {
		var zeroT T
		return zeroT, fmt.Errorf("queue is empty")
	}
	keys := maps.Keys(s.queue)
	top := keys[0]
	delete(s.queue, top)
	return top, nil
}

func (s *Session[T]) Remove(t T) {
	s.Lock()
	defer s.Unlock()
	delete(s.queue, t)
}

func (s *Session[T]) CanMatch() bool {
	s.Lock()
	defer s.Unlock()
	return len(s.queue) >= 2
}
