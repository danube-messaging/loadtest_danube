package metrics

import (
	"sort"
	"testing"
)

func TestPercentilesEmpty(t *testing.T) {
	p50, p95, p99, pmax := percentiles(nil)
	if p50 != 0 || p95 != 0 || p99 != 0 || pmax != 0 {
		t.Fatalf("expected zeros, got p50=%v p95=%v p99=%v pmax=%v", p50, p95, p99, pmax)
	}
}

func TestPercentilesSingle(t *testing.T) {
	in := []float64{42}
	p50, p95, p99, pmax := percentiles(in)
	if p50 != 42 || p95 != 42 || p99 != 42 || pmax != 42 {
		t.Fatalf("unexpected: %v %v %v %v", p50, p95, p99, pmax)
	}
}

func TestPercentilesTypical(t *testing.T) {
	in := []float64{10, 1, 5, 7, 50, 20, 3, 2, 100, 8}
	sort.Float64s(in) // percentiles expects sorted input
	p50, p95, p99, pmax := percentiles(in)
	if pmax != 100 {
		t.Fatalf("pmax got %v want 100", pmax)
	}
	if p50 <= 7 || p50 > 10 { // rough median range for this set
		t.Fatalf("p50 out of expected range: %v", p50)
	}
	if p95 < p50 || p99 < p95 {
		t.Fatalf("monotonic percentile order violated: p50=%v p95=%v p99=%v", p50, p95, p99)
	}
}
