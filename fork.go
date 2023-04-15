package fork

import (
	"sync"
	"sync/atomic"
)

const defaultParallelism = 1

// Fork ...
type Fork[IN, OUT any] interface {

	// Concurrency sets the number of goroutines to be used
	Concurrency(numGoroutines int) Fork[IN, OUT]

	// JoinSlice returns the processing results as a slice
	JoinSlice(func(input IN) (output OUT, terminate bool)) []OUT

	// JoinChan returns the processing results as a chan
	JoinChan(func(input IN) (output OUT, terminate bool)) <-chan OUT
}

type fork[IN, OUT any] struct {
	routineCount int
	iterator     Iterator[IN]
}

func (f *fork[IN, OUT]) numGoroutines() int {
	if f.routineCount < defaultParallelism {
		return defaultParallelism
	}
	return f.routineCount
}

// Concurrency ...
func (f *fork[IN, OUT]) Concurrency(numGoroutines int) Fork[IN, OUT] {
	f.routineCount = numGoroutines
	return f
}

// JoinSlice ...
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

// JoinChan ...
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
