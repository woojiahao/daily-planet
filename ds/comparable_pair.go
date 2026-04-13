package ds

type ComparablePair[A comparable, B comparable] struct {
	First  A
	Second B
}

func NewComparablePair[A comparable, B comparable](first A, second B) *ComparablePair[A, B] {
	return &ComparablePair[A, B]{First: first, Second: second}
}
