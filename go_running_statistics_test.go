package runningstats

import (
	"math"
	"testing"
)

const eps = 1e-9

func almostEqual(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > eps {
		t.Fatalf("got %v, want %v (diff %v)", got, want, math.Abs(got-want))
	}
}

func TestEmptyStream(t *testing.T) {
	r := New()
	if r.Count() != 0 {
		t.Fatalf("Count = %d, want 0", r.Count())
	}
	almostEqual(t, r.Mean(), 0)
	almostEqual(t, r.Variance(), 0)
	almostEqual(t, r.PopulationVariance(), 0)
	almostEqual(t, r.StdDev(), 0)
}

func TestZeroValueIsUsable(t *testing.T) {
	var r RunningStats
	r.Add(4)
	r.Add(8)
	if r.Count() != 2 {
		t.Fatalf("Count = %d, want 2", r.Count())
	}
	almostEqual(t, r.Mean(), 6)
}

func TestSingleValue(t *testing.T) {
	r := New()
	r.Add(42)
	almostEqual(t, r.Mean(), 42)
	// Sample variance is undefined for n < 2; we define it as 0.
	almostEqual(t, r.Variance(), 0)
	almostEqual(t, r.PopulationVariance(), 0)
}

func TestKnownValues(t *testing.T) {
	r := New()
	r.AddAll([]float64{2, 4, 4, 4, 5, 5, 7, 9})
	if r.Count() != 8 {
		t.Fatalf("Count = %d, want 8", r.Count())
	}
	almostEqual(t, r.Mean(), 5)
	// Population variance of this classic data set is 4.
	almostEqual(t, r.PopulationVariance(), 4)
	almostEqual(t, r.PopulationStdDev(), 2)
	// Sample variance is 32/7.
	almostEqual(t, r.Variance(), 32.0/7.0)
}

func TestLargeOffset(t *testing.T) {
	// A naive sum-of-squares implementation loses precision badly here.
	// Welford's algorithm should stay accurate.
	const base = 1e9
	r := New()
	r.Add(base + 1)
	r.Add(base + 2)
	r.Add(base + 3)
	r.Add(base + 4)
	almostEqual(t, r.Mean(), base+2.5)
	// Sample variance of {1,2,3,4} is 5/3.
	almostEqual(t, r.Variance(), 5.0/3.0)
}

func TestNegativeValues(t *testing.T) {
	r := New()
	r.AddAll([]float64{-1, -2, -3, -4})
	almostEqual(t, r.Mean(), -2.5)
	almostEqual(t, r.Variance(), 5.0/3.0)
}

func TestReset(t *testing.T) {
	r := New()
	r.AddAll([]float64{1, 2, 3})
	r.Reset()
	if r.Count() != 0 {
		t.Fatalf("Count after Reset = %d, want 0", r.Count())
	}
	almostEqual(t, r.Mean(), 0)
	almostEqual(t, r.Variance(), 0)
}

func TestMergeIntoEmpty(t *testing.T) {
	a := New()
	b := New()
	b.AddAll([]float64{1, 2, 3, 4})
	a.Merge(b)
	if a.Count() != 4 {
		t.Fatalf("Count = %d, want 4", a.Count())
	}
	almostEqual(t, a.Mean(), 2.5)
	almostEqual(t, a.Variance(), 5.0/3.0)
}

func TestMergeTwoNonEmpty(t *testing.T) {
	a := New()
	a.AddAll([]float64{1, 2, 3})
	b := New()
	b.AddAll([]float64{4, 5, 6})
	a.Merge(b)
	if a.Count() != 6 {
		t.Fatalf("Count = %d, want 6", a.Count())
	}
	almostEqual(t, a.Mean(), 3.5)
	// Sample variance of {1..6} is 3.5.
	almostEqual(t, a.Variance(), 3.5)
}

func TestMergeNilAndEmpty(t *testing.T) {
	a := New()
	a.Add(10)
	a.Merge(nil)
	a.Merge(New())
	if a.Count() != 1 {
		t.Fatalf("Count = %d, want 1", a.Count())
	}
	almostEqual(t, a.Mean(), 10)
}

func TestMergeMatchesSequential(t *testing.T) {
	// Merging two halves must produce the same statistics as adding all
	// values one at a time.
	left := New()
	left.AddAll([]float64{3, 1, 4, 1, 5})
	right := New()
	right.AddAll([]float64{9, 2, 6, 5, 3})
	left.Merge(right)

	seq := New()
	seq.AddAll([]float64{3, 1, 4, 1, 5, 9, 2, 6, 5, 3})

	almostEqual(t, left.Mean(), seq.Mean())
	almostEqual(t, left.Variance(), seq.Variance())
	almostEqual(t, left.PopulationVariance(), seq.PopulationVariance())
}

func TestPopulationVsSample(t *testing.T) {
	r := New()
	r.AddAll([]float64{1, 2, 3, 4})
	// Population variance of {1,2,3,4} is 1.25; sample is 5/3.
	almostEqual(t, r.PopulationVariance(), 1.25)
	almostEqual(t, r.Variance(), 5.0/3.0)
	almostEqual(t, r.StdDev(), math.Sqrt(5.0/3.0))
	almostEqual(t, r.PopulationStdDev(), math.Sqrt(1.25))
}
