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

func Slice[IN, OUT any](input []IN) Fork[IN, OUT] {
	inputChannel := make(chan IN, len(input))
	for _, value := range input {
		inputChannel <- value
	}
	close(inputChannel)
	return &fork[IN, OUT]{
		input:       inputChannel,
		parallelism: defaultParallelism,
	}
}

func Chan[IN, OUT any](input <-chan IN) Fork[IN, OUT] {
	return &fork[IN, OUT]{
		input:       input,
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
		input:       inputChannel,
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
		input:       inputChannel,
		parallelism: defaultParallelism,
	}
}

type fork[IN, OUT any] struct {
	parallelism int
	input       <-chan IN
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
				value, hasMore := <-f.input
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
					value, hasMore := <-f.input
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
