// Package runningstats computes mean, variance, and standard deviation of a
// stream of numbers in a single pass.
//
// The package name is runningstats rather than go_running_statistics because Go
// package identifiers cannot contain underscores by convention and the import
// path already carries the module name.
package runningstats

import "math"

// RunningStats accumulates the count, mean, and sum of squared deviations from
// the mean (M2) for a stream of float64 values.
//
// It uses Welford's online algorithm, which avoids the catastrophic cancellation
// that a naive sum-of-squares approach suffers when the mean is large relative
// to the spread of the data. The zero value is ready to use and represents an
// empty stream.
type RunningStats struct {
	count int
	mean  float64
	m2    float64
}

// New returns an empty RunningStats. It is equivalent to declaring the zero
// value, and exists so that callers who prefer an explicit constructor have one.
func New() *RunningStats {
	return &RunningStats{}
}

// Add incorporates x into the running statistics.
func (r *RunningStats) Add(x float64) {
	r.count++
	delta := x - r.mean
	r.mean += delta / float64(r.count)
	delta2 := x - r.mean
	r.m2 += delta * delta2
}

// AddAll incorporates every value in xs into the running statistics.
func (r *RunningStats) AddAll(xs []float64) {
	for _, x := range xs {
		r.Add(x)
	}
}

// Count returns the number of values added so far.
func (r *RunningStats) Count() int {
	return r.count
}

// Mean returns the arithmetic mean of the values added so far.
//
// It returns 0 when no values have been added, so callers that need to
// distinguish an empty stream from a stream whose mean is exactly zero should
// check Count first.
func (r *RunningStats) Mean() float64 {
	if r.count == 0 {
		return 0
	}
	return r.mean
}

// Variance returns the sample variance (dividing by n-1) of the values added
// so far.
//
// Sample variance is the default because streaming statistics are usually
// applied to a sample drawn from a larger population. It returns 0 when fewer
// than two values have been added, since sample variance is undefined there.
func (r *RunningStats) Variance() float64 {
	if r.count < 2 {
		return 0
	}
	return r.m2 / float64(r.count-1)
}

// PopulationVariance returns the population variance (dividing by n) of the
// values added so far. It returns 0 when no values have been added.
func (r *RunningStats) PopulationVariance() float64 {
	if r.count == 0 {
		return 0
	}
	return r.m2 / float64(r.count)
}

// StdDev returns the sample standard deviation of the values added so far.
func (r *RunningStats) StdDev() float64 {
	return math.Sqrt(r.Variance())
}

// PopulationStdDev returns the population standard deviation of the values
// added so far.
func (r *RunningStats) PopulationStdDev() float64 {
	return math.Sqrt(r.PopulationVariance())
}

// Merge combines another RunningStats into r, as if all of other's values had
// been added to r in sequence. It is exact for the mean and for M2, using the
// parallel form of Welford's update.
//
// Merging an empty other is a no-op. Merging into an empty receiver copies
// other's state.
func (r *RunningStats) Merge(other *RunningStats) {
	if other == nil || other.count == 0 {
		return
	}
	if r.count == 0 {
		r.count = other.count
		r.mean = other.mean
		r.m2 = other.m2
		return
	}

	nA := float64(r.count)
	nB := float64(other.count)
	n := nA + nB
	delta := other.mean - r.mean

	r.mean = (nA*r.mean + nB*other.mean) / n
	r.m2 += other.m2 + delta*delta*nA*nB/n
	r.count += other.count
}

// Reset clears the accumulator so it can be reused for a new stream.
func (r *RunningStats) Reset() {
	r.count = 0
	r.mean = 0
	r.m2 = 0
}
