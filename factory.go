package fork

func Iter[IN, OUT any](input Iterator[IN]) Fork[IN, OUT] {
	return &fork[IN, OUT]{
		iterator: input,
	}
}

// Slice source
func Slice[IN, OUT any](input []IN) Fork[IN, OUT] {
	return &fork[IN, OUT]{
		iterator: &sliceIterator[IN]{
			slice: input,
		},
	}
}

// Chan source
func Chan[IN, OUT any](input <-chan IN) Fork[IN, OUT] {
	return &fork[IN, OUT]{
		iterator: &chanIterator[IN]{
			channel: input,
		},
	}
}

// Keys of map source
func Keys[K comparable, V any, OUT any](input map[K]V) Fork[K, OUT] {
	return &fork[K, OUT]{
		iterator: &mapKeyIterator[K, V]{
			sourceMap: input,
			keys:      make(chan K),
			done:      make(chan struct{}),
		},
	}
}

// Values of map source
func Values[K comparable, V any, OUT any](input map[K]V) Fork[V, OUT] {
	return &fork[V, OUT]{
		iterator: &mapValueIterator[K, V]{
			sourceMap: input,
			values:    make(chan V),
			done:      make(chan struct{}),
		},
	}
}
