package structs

import "iter"

type Set[T comparable] map[T]any

func NewSet[T comparable]() Set[T] {
	return make(Set[T])
}

func (s *Set[T]) Add(value T) {
	(*s)[value] = struct{}{}
}

func (s *Set[T]) Remove(value T) {
	delete(*s, value)
}

func (s Set[T]) Includes(value T) bool {
	_, ok := s[value]

	return ok
}

func (s Set[T]) Len() int {
	return len(s)
}

func (s Set[T]) GetItems() iter.Seq[T] {
	return func(yield func(T) bool) {
		for item := range s {
			if !yield(item) {
				return
			}
		}
	}
}

func (s Set[T]) Copy() Set[T] {
	result := NewSet[T]()

	for item := range s.GetItems() {
		result.Add(item)
	}

	return result
}
