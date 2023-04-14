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
squares := Slice[int, int](input).
    Parallelism(3).
    ToChan(func(input int) (int, bool) {
        return input*input, false
    })
```

Fork from a chan and get the result as a slice
```go
input := make(chan int)
squares := Slice[int, int](input).
    Parallelism(3).
    ToSlice(func(input int) (int, bool) {
        return input*input, false
    })
```

Signal an early exit
```go
input := []int{1, 2, 3}
squares := Slice[int, int](input).
    Parallelism(3).
    ToChan(func(input int) (int, bool) {
        return input*input, true // This signals that we want to stop
    })
```

Fork from map values to a channel
```go
input := map[string]int{
    "zero": 0,
    "one":  1,
    "two":  2,
}
valueSquares := Values[string, int, int](input).
    ToChan(func(input int) (int, bool) {
        return input*input, false
    })
```
Fork from map keys to a channel
```go
input := map[int]string{
    0: "zero",
    1: "one",
    2: "two",
}
keySquares := Keys[int, string, int](input).
    ToSlice(func(input int) (int, bool) {
        return input*input, false
    })
```