# go_running_statistics

Computes the mean, sample variance, sample standard deviation, population
variance, and population standard deviation of a stream of `float64` values in
a single pass, using Welford's online algorithm.

```go
package main

import (
	"fmt"

	runningstats "github.com/example/go_running_statistics"
)

func main() {
	r := runningstats.New()
	r.Add(2)
	r.Add(4)
	r.Add(4)
	r.Add(4)
	r.Add(5)
	r.Add(5)
	r.Add(7)
	r.Add(9)

	fmt.Println(r.Count())              // 8
	fmt.Println(r.Mean())               // 5
	fmt.Println(r.Variance())           // 4.571428571428571 (sample, n-1)
	fmt.Println(r.PopulationVariance()) // 4
	fmt.Println(r.StdDev())             // 2.138089935299395
}
```

## Why this exists

The obvious way to compute variance in a stream is to accumulate `sum` and
`sum of squares`, then evaluate `E[x^2] - E[x]^2` at the end. That formula is
mathematically correct but numerically terrible: when the values are large
relative to their spread (timestamps, sensor offsets, prices around a fixed
baseline) the two large terms cancel and you can lose most of your significant
digits, sometimes ending up with a negative variance.

Welford's algorithm updates the mean and the sum of squared deviations from the
mean incrementally, so the quantities being subtracted are always small. The
trade-off is one extra division per value compared to the naive accumulator, and
a slightly less obvious update rule. For streaming statistics that is almost
always the right trade.

## API

- `type RunningStats struct` — accumulator. The zero value is ready to use and
  represents an empty stream.
- `func New() *RunningStats` — explicit constructor; equivalent to the zero
  value.
- `func (r *RunningStats) Add(x float64)`
- `func (r *RunningStats) AddAll(xs []float64)`
- `func (r *RunningStats) Count() int`
- `func (r *RunningStats) Mean() float64`
- `func (r *RunningStats) Variance() float64` — sample variance, divides by
  `n-1`.
- `func (r *RunningStats) PopulationVariance() float64` — divides by `n`.
- `func (r *RunningStats) StdDev() float64`
- `func (r *RunningStats) PopulationStdDev() float64`
- `func (r *RunningStats) Merge(other *RunningStats)`
- `func (r *RunningStats) Reset()`

## The awkward edge

There is no way to distinguish "no values added yet" from "values added whose
mean happens to be zero" by looking at `Mean()` alone — it returns `0` in both
cases. Check `Count()` first if that distinction matters to you. The same
applies to the variance accessors: they return `0` when the stream is too short
for the statistic to be defined (`n == 0` for population, `n < 2` for sample),
rather than panicking or returning `NaN`.

`Merge` is exact for the mean and for the sum of squared deviations; it does not
re-round the merged values, so a merge of two accumulators may differ from a
sequential `Add` of the same values by a few ULPs.
