package fork

import (
	"sync"
	"sync/atomic"
)

const defaultParallelism = 1

// Fork ...
type Fork[IN, OUT any] interface {

	// Parallelism sets the number of goroutines to be used
	Parallelism(parallelism int) Fork[IN, OUT]

	// JoinSlice returns the processing results as a slice
	JoinSlice(func(input IN) (output OUT, terminate bool)) []OUT

	// JoinChan returns the processing results as a chan
	JoinChan(func(input IN) (output OUT, terminate bool)) <-chan OUT
}

type fork[IN, OUT any] struct {
	parallelism int
	iterator    Iterator[IN]
}

func (f *fork[IN, OUT]) numGoroutines() int {
	if f.parallelism < defaultParallelism {
		return defaultParallelism
	}
	return f.parallelism
}

// Parallelism setter
func (f *fork[IN, OUT]) Parallelism(parallelism int) Fork[IN, OUT] {
	f.parallelism = parallelism
	return f
}

// JoinSlice transforms and returns the results as a slice
func (f *fork[IN, OUT]) JoinSlice(transformer func(_ IN) (OUT, bool)) []OUT {
	var (
		results   []OUT
		lock      sync.Mutex
		wg        sync.WaitGroup
		isRunning atomic.Bool
	)
	wg.Add(f.numGoroutines())
	isRunning.Store(true)
	f.iterator.Open()
	defer f.iterator.Close()
	for i := 0; i < f.numGoroutines(); i++ {
		go func() {
			defer wg.Done()
			for isRunning.Load() {
				value, hasMore := f.iterator.Next()
				if !hasMore {
					break
				}
				result, terminate := transformer(value)
				if terminate {
					isRunning.Store(false)
				}
				lock.Lock()
				results = append(results, result)
				lock.Unlock()
			}
		}()
	}

	wg.Wait()

	return results
}

// JoinChan transforms and returns the results as a channel
func (f *fork[IN, OUT]) JoinChan(transformer func(_ IN) (OUT, bool)) <-chan OUT {
	results := make(chan OUT)
	go func() {
		var (
			wg        = sync.WaitGroup{}
			isRunning = atomic.Bool{}
		)
		wg.Add(f.numGoroutines())
		isRunning.Store(true)
		f.iterator.Open()
		defer close(results)
		defer f.iterator.Close()
		defer wg.Wait()

		for i := 0; i < f.numGoroutines(); i++ {
			go func() {
				for isRunning.Load() {
					value, hasMore := f.iterator.Next()
					if !hasMore {
						break
					}
					result, terminate := transformer(value)
					if terminate {
						isRunning.Store(false)
						break
					}
					results <- result
				}
				wg.Done()
			}()
		}
	}()
	return results
}
