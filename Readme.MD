## Fork

### A micro library for in-memory fork join
This seems to be a recurring pattern that results in lots of copy/paste code

#### Supported sources
* Slice
* Chan
* Map keys
* Map values
* Iterator
#### Supported destinations
* Channel
* Slice

#### Examples

Fork from slice with parallelism of 3 and join results into a channel
```go
input := []string{"dog", "cat", "moose"}
lengths := fork.
	Slice[string, int](input).
	Parallelism(3).
	JoinChan(func(value string) (int, bool) {
		return len(value), false
	})
```

Fork from a channel and get the result as a slice
```go
input := make(chan int)
squares := fork.
	Chan[int, int](input).
	Parallelism(3).
	JoinSlice(func(value int) (int, bool) {
		return value*value, false
	})
```

Signal an early exit
```go
input := []int{1, 2, 3}
squares := fork.
	Slice[int, int](input).
	Parallelism(3).
	JoinChan(func(value int) (int, bool) {
		// This signals that we want to stop. Maybe be triggered by an error... who knows
		return value*value, true 
	})
```

Fork from map values to a channel
```go
input := map[string]int{
    "zero": 0,
    "one":  1,
    "two":  2,
}
valueSquares := fork.
	Values[string, int, int](input).
	JoinChan(func(value int) (int, bool) {
		return value*value, false
	})
```
Fork from map keys to a channel
```go
input := map[int]string{
    0: "zero",
    1: "one",
    2: "two",
}
keySquares := fork.
	Keys[int, string, int](input).
	JoinSlice(func(value int) (int, bool) {
		return value*value, false
	})
```

Fork from a custom iterator
```go
type tickerIterator struct {
    ticker   *time.Ticker
    deadline time.Time
}

func (c *tickerIterator) Open() {
}

func (c *tickerIterator) Close() {
    c.ticker.Stop()
}

func (c *tickerIterator) Next() (*time.Time, bool) {
    t, ok := <-c.ticker.C
    if time.Now().After(c.deadline) {
        return nil, false
    }
    return &t, ok
}

source := &tickerIterator{
    ticker:   time.NewTicker(time.Second),
    deadline: time.Now().Add(time.Second * 5),
}
output := fork.Iter[*time.Time, int64](source).
    JoinSlice(func(input *time.Time) (output int64, terminate bool) {
        return input.UnixMilli(), false
    })
// output should contain the unix millis from the ticker generated time.Time
fmt.Println(output)
```