## Fork

### A micro library for in-memory fork join
This seems to be a recurring pattern that results in lots of copy/paste code

#### Supported flows
* Slice to Chan, Slice
* Chan to Chan, Slice
* Map keys to Chan, Slice
* Map values to Chan, Slice

#### Examples

Fork from slice with parallelism of 3 and join results into a channel
```go
input := []int{1, 2, 3}
squares := fork.Slice[int, int](input).
    Parallelism(3).
    ToChan(func(value int) (int, bool) {
        return value*value, false
    })
```

Fork from a channel and get the result as a slice
```go
input := make(chan int)
squares := fork.Chan[int, int](input).
    Parallelism(3).
    ToSlice(func(value int) (int, bool) {
        return value*value, false
    })
```

Signal an early exit
```go
input := []int{1, 2, 3}
squares := fork.Slice[int, int](input).
    Parallelism(3).
    ToChan(func(value int) (int, bool) {
        return value*value, true // This signals that we want to stop. Maybe be triggered by an error... who knows
    })
```

Fork from map values to a channel
```go
input := map[string]int{
    "zero": 0,
    "one":  1,
    "two":  2,
}
valueSquares := fork.Values[string, int, int](input).
    ToChan(func(value int) (int, bool) {
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
keySquares := fork.Keys[int, string, int](input).
    ToSlice(func(value int) (int, bool) {
        return value*value, false
    })
```