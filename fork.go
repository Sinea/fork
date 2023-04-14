package fork

import (
	"sync"
	"sync/atomic"
)

const defaultParallelism = 1

// Fork object
type Fork[IN, OUT any] interface {

	// Parallelism sets the number of goroutines to be used
	Parallelism(parallelism int) Fork[IN, OUT]

	// ToSlice returns the processing results as a slice
	ToSlice(func(_ IN) (OUT, bool)) []OUT

	// ToChan returns the processing results as a chan
	ToChan(func(_ IN) (OUT, bool)) <-chan OUT
}

// Slice source
func Slice[IN, OUT any](input []IN) Fork[IN, OUT] {
	return &fork[IN, OUT]{
		iter: &sliceIterator[IN]{
			slice: input,
		},
		parallelism: defaultParallelism,
	}
}

// Chan souce
func Chan[IN, OUT any](input <-chan IN) Fork[IN, OUT] {
	return &fork[IN, OUT]{
		iter: &chanIterator[IN]{
			channel: input,
		},
		parallelism: defaultParallelism,
	}
}

// Keys of map source
func Keys[K comparable, V any, OUT any](input map[K]V) Fork[K, OUT] {
	return &fork[K, OUT]{
		iter: &mapKeyIterator[K, V]{
			sourceMap: input,
			keys:      make(chan K),
			done:      make(chan struct{}),
		},
		parallelism: defaultParallelism,
	}
}

// Values of map source
func Values[K comparable, V any, OUT any](input map[K]V) Fork[V, OUT] {
	return &fork[V, OUT]{
		iter: &mapValueIterator[K, V]{
			sourceMap: input,
			values:    make(chan V),
			done:      make(chan struct{}),
		},
		parallelism: defaultParallelism,
	}
}

type fork[IN, OUT any] struct {
	parallelism int
	iter        Iterator[IN]
}

// Parallelism setter
func (f *fork[IN, OUT]) Parallelism(parallelism int) Fork[IN, OUT] {
	if parallelism < defaultParallelism {
		parallelism = defaultParallelism
	}
	f.parallelism = parallelism
	return f
}

// ToSlice transforms and returns the results as a slice
func (f *fork[IN, OUT]) ToSlice(transformer func(_ IN) (OUT, bool)) []OUT {
	wg := sync.WaitGroup{}
	wg.Add(f.parallelism)
	var (
		results []OUT
		lock    sync.Mutex
	)
	isRunning := atomic.Bool{}
	isRunning.Store(true)
	f.iter.Open()
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
	f.iter.Close()
	return results
}

// ToChan transforms and returns the results as a channel
func (f *fork[IN, OUT]) ToChan(transformer func(_ IN) (OUT, bool)) <-chan OUT {
	results := make(chan OUT)
	go func() {
		wg := sync.WaitGroup{}
		wg.Add(f.parallelism)
		isRunning := atomic.Bool{}
		isRunning.Store(true)
		f.iter.Open()
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
		f.iter.Close()
		close(results)
	}()
	return results
}
