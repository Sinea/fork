package fork

import "sync"

// Iterator returns values or false if n more values are present
type Iterator[T any] interface {
	// Open the iterator
	Open()

	// Close the iterator
	Close()

	// Next value
	Next() (T, bool)
}

type chanIterator[T any] struct {
	channel <-chan T
}

func (c *chanIterator[T]) Open() {

}

func (c *chanIterator[T]) Close() {

}

func (c *chanIterator[T]) Next() (T, bool) {
	value, ok := <-c.channel
	return value, ok
}

type sliceIterator[T any] struct {
	sync.Mutex
	slice []T
	index int
}

func (s *sliceIterator[T]) Open() {

}

func (s *sliceIterator[T]) Close() {

}

type mapKeyIterator[K comparable, V any] struct {
	sourceMap map[K]V
	keys      chan K
	done      chan struct{}
}

func (m *mapKeyIterator[K, V]) Open() {
	go func() {
		defer close(m.keys)
	Loop:
		for k := range m.sourceMap {
			select {
			case m.keys <- k:
			case <-m.done:
				break Loop
			}
		}
	}()
}

func (m *mapKeyIterator[K, V]) Close() {
	close(m.done)
}

func (m *mapKeyIterator[K, V]) Next() (K, bool) {
	value, ok := <-m.keys
	return value, ok
}

type mapValueIterator[K comparable, V any] struct {
	sourceMap map[K]V
	values    chan V
	done      chan struct{}
}

func (m *mapValueIterator[K, V]) Open() {
	go func() {
		defer close(m.values)
	Loop:
		for _, v := range m.sourceMap {
			select {
			case m.values <- v:
			case <-m.done:
				break Loop
			}
		}
	}()
}

func (m *mapValueIterator[K, V]) Close() {
	close(m.done)
}

func (m *mapValueIterator[K, V]) Next() (V, bool) {
	value, ok := <-m.values
	return value, ok
}

// Next item in line
func (s *sliceIterator[T]) Next() (T, bool) {
	s.Lock()
	defer s.Unlock()
	if s.index >= len(s.slice) {
		var t T
		return t, false
	}
	s.index++
	return s.slice[s.index-1], true
}
