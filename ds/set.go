// Package ds contains the data structures implemented
package ds

type Set[T comparable] struct {
	elements map[T]struct{}
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		elements: make(map[T]struct{}),
	}
}

func FromArray[T comparable](arr []T) *Set[T] {
	set := NewSet[T]()
	for _, a := range arr {
		set.Add(a)
	}
	return set
}

func (s *Set[T]) Add(value T) {
	s.elements[value] = struct{}{}
}

func (s *Set[T]) Remove(value T) {
	delete(s.elements, value)
}

func (s Set[T]) Contains(value T) bool {
	_, exists := s.elements[value]
	return exists
}

func (s Set[T]) Size() int {
	return len(s.elements)
}

func (s Set[T]) Values() []T {
	values := make([]T, 0, s.Size())
	for v := range s.elements {
		values = append(values, v)
	}
	return values
}
