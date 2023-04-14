package fork

import "sync"

// Iterator returns values or false if no more values are present
type Iterator[T any] interface {
	// Open the iterator
	Open()

	// Close the iterator
	Close()

	// Next value and false if there are no more values
	Next() (T, bool)
}

type chanIterator[T any] struct {
	channel <-chan T
}

// Open ...
func (c *chanIterator[T]) Open() {

}

// Close ...
func (c *chanIterator[T]) Close() {

}

// Next ...
func (c *chanIterator[T]) Next() (T, bool) {
	value, ok := <-c.channel
	return value, ok
}

type sliceIterator[T any] struct {
	sync.Mutex
	slice []T
	index int
}

// Open ...
func (s *sliceIterator[T]) Open() {

}

// Close ...
func (s *sliceIterator[T]) Close() {

}

// Next ...
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

type mapKeyIterator[K comparable, V any] struct {
	sourceMap map[K]V
	keys      chan K
	done      chan struct{}
	closeOnce sync.Once
}

// Open ...
func (m *mapKeyIterator[K, V]) Open() {
	go func() {
		defer close(m.keys)
		defer m.Close()
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

// Close ...
func (m *mapKeyIterator[K, V]) Close() {
	m.closeOnce.Do(func() {
		close(m.done)
	})
}

// Next ...
func (m *mapKeyIterator[K, V]) Next() (K, bool) {
	value, ok := <-m.keys
	return value, ok
}

type mapValueIterator[K comparable, V any] struct {
	sourceMap map[K]V
	values    chan V
	done      chan struct{}
	closeOnce sync.Once
}

// Open ...
func (m *mapValueIterator[K, V]) Open() {
	go func() {
		defer close(m.values)
		defer m.Close()
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

// Close ...
func (m *mapValueIterator[K, V]) Close() {
	m.closeOnce.Do(func() {
		close(m.done)
	})
}

// Next ...
func (m *mapValueIterator[K, V]) Next() (V, bool) {
	value, ok := <-m.values
	return value, ok
}
