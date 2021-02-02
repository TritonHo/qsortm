
qsortm = qsort with multithreading

A 100% compatible drop-in replacement of Sort() and Slice() in "sort"

# Benchmark


Raw Result (on AMD Ryzen 3600, 6c12t, 1M records)

|               | ns/op        |
| ------------- | ------------ |
| qsortm Sort   | 43.92 ns/op  |
| qsortm Slice  | 32.65 ns/op  |
| std Sort      | 193.57 ns/op |
| std Slice     | 168.86 ns/op |


# Usage

100% compatibility with standard lib

Replace

```go
import "sort"

sort.Sort(slice)
sort.Slice(slice, lessFn)
```

with

```go
import "github.com/TritonHo/qsortm"

qsortm.Sort(slice)
qsortm.Slice(slice, lessFn)
```
