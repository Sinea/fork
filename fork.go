package fork

import (
	"sync"
	"sync/atomic"
)

const defaultParallelism = 1

// Iterator returns values or false if n more values are present
type Iterator[T any] interface {
	Next() (T, bool)
}

type chanIterator[T any] struct {
	channel <-chan T
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

// Fork object
type Fork[IN, OUT any] interface {

	// Parallelism sets the number of goroutines to be used
	Parallelism(parallelism int) Fork[IN, OUT]

	// ToSlice returns the processing results as a slice
	ToSlice(func(_ IN) (OUT, bool)) []OUT

	// ToChan returns the processing results as a chan
	ToChan(func(_ IN) (OUT, bool)) <-chan OUT
}

func Slice[IN, OUT any](input []IN) Fork[IN, OUT] {
	return &fork[IN, OUT]{
		iter: &sliceIterator[IN]{
			slice: input,
		},
		parallelism: defaultParallelism,
	}
}

func Chan[IN, OUT any](input <-chan IN) Fork[IN, OUT] {
	return &fork[IN, OUT]{
		iter: &chanIterator[IN]{
			channel: input,
		},
		parallelism: defaultParallelism,
	}
}

func Keys[K comparable, V any, OUT any](m map[K]V) Fork[K, OUT] {
	inputChannel := make(chan K, len(m))
	for key := range m {
		inputChannel <- key
	}
	close(inputChannel)
	return &fork[K, OUT]{
		iter: &chanIterator[K]{
			channel: inputChannel,
		},
		parallelism: defaultParallelism,
	}
}

func Values[K comparable, V any, OUT any](m map[K]V) Fork[V, OUT] {
	inputChannel := make(chan V, len(m))
	for _, value := range m {
		inputChannel <- value
	}
	close(inputChannel)
	return &fork[V, OUT]{
		iter: &chanIterator[V]{
			channel: inputChannel,
		},
		parallelism: defaultParallelism,
	}
}

type fork[IN, OUT any] struct {
	parallelism int
	iter        Iterator[IN]
}

func (f *fork[IN, OUT]) Parallelism(parallelism int) Fork[IN, OUT] {
	if parallelism < defaultParallelism {
		parallelism = defaultParallelism
	}
	f.parallelism = parallelism
	return f
}

func (f *fork[IN, OUT]) ToSlice(transformer func(_ IN) (OUT, bool)) []OUT {
	wg := sync.WaitGroup{}
	wg.Add(f.parallelism)
	var (
		results []OUT
		lock    sync.Mutex
	)
	isRunning := atomic.Bool{}
	isRunning.Store(true)
	for i := 0; i < f.parallelism; i++ {
		go func() {
			for isRunning.Load() {
				value, hasMore := f.iter.Next()
				if !hasMore {
					break
				}
				result, stopIteration := transformer(value)
				if stopIteration {
					isRunning.Store(false)
				}
				lock.Lock()
				results = append(results, result)
				lock.Unlock()
			}
			wg.Done()
		}()
	}

	wg.Wait()

	return results
}

func (f *fork[IN, OUT]) ToChan(transformer func(_ IN) (OUT, bool)) <-chan OUT {
	results := make(chan OUT)
	go func() {
		wg := sync.WaitGroup{}
		wg.Add(f.parallelism)
		isRunning := atomic.Bool{}
		isRunning.Store(true)
		for i := 0; i < f.parallelism; i++ {
			go func() {
				for isRunning.Load() {
					value, hasMore := f.iter.Next()
					if !hasMore {
						break
					}
					result, stopIteration := transformer(value)
					if stopIteration {
						isRunning.Store(false)
						break
					}
					results <- result
				}
				wg.Done()
			}()
		}

		wg.Wait()
		close(results)
	}()
	return results
}
