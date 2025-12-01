package city

import "iter"

type Set[T comparable] struct {
	items map[T]any
}

func NewSet[T comparable]() Set[T] {
	return Set[T]{items: make(map[T]any)}
}

func (s *Set[T]) Add(value T) {
	s.items[value] = struct{}{}
}

func (s *Set[T]) Remove(value T) {
	delete(s.items, value)
}

func (s *Set[T]) Includes(value T) bool {
	_, ok := s.items[value]

	return ok
}

func (s *Set[T]) Len() int {
	return len(s.items)
}

func (s *Set[T]) GetItems() iter.Seq[T] {
	return func(yield func(T) bool) {
		for item := range s.items {
			if !yield(item) {
				return
			}
		}
	}
}
