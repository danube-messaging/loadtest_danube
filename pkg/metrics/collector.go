package metrics

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Collector struct {
	Start time.Time

	MessagesSent     atomic.Uint64
	MessagesReceived atomic.Uint64
	Errors           atomic.Uint64

	mu        sync.Mutex
	latencies []float64 // milliseconds
}

func NewCollector() *Collector {
	return &Collector{Start: time.Now()}
}

func (c *Collector) IncSent(n uint64)     { c.MessagesSent.Add(n) }
func (c *Collector) IncReceived(n uint64) { c.MessagesReceived.Add(n) }
func (c *Collector) IncError(n uint64)    { c.Errors.Add(n) }

// RecordLatency adds an end-to-end latency sample in milliseconds.
func (c *Collector) RecordLatency(ms float64) {
	c.mu.Lock()
	c.latencies = append(c.latencies, ms)
	c.mu.Unlock()
}

func (c *Collector) Snapshot() Snapshot {
	elapsed := time.Since(c.Start).Seconds()
	sent := c.MessagesSent.Load()
	recv := c.MessagesReceived.Load()
	errs := c.Errors.Load()
	// copy latencies to avoid holding lock during sort
	c.mu.Lock()
	lcopy := append([]float64(nil), c.latencies...)
	c.mu.Unlock()
	sort.Float64s(lcopy)
	p50, p95, p99, pmax := percentiles(lcopy)
	return Snapshot{
		ElapsedSec:       elapsed,
		MessagesSent:     sent,
		MessagesReceived: recv,
		Errors:           errs,
		ThroughputSent:   rate(sent, elapsed),
		ThroughputRecv:   rate(recv, elapsed),
		LatencyP50Ms:     p50,
		LatencyP95Ms:     p95,
		LatencyP99Ms:     p99,
		LatencyMaxMs:     pmax,
		LatencySamples:   len(lcopy),
	}
}

type Snapshot struct {
	ElapsedSec       float64
	MessagesSent     uint64
	MessagesReceived uint64
	Errors           uint64
	ThroughputSent   float64
	ThroughputRecv   float64
	LatencyP50Ms     float64
	LatencyP95Ms     float64
	LatencyP99Ms     float64
	LatencyMaxMs     float64
	LatencySamples   int
}

func rate(total uint64, elapsedSec float64) float64 {
	if elapsedSec <= 0 {
		return 0
	}
	return float64(total) / elapsedSec
}

func percentiles(sorted []float64) (p50, p95, p99, pmax float64) {
	n := len(sorted)
	if n == 0 {
		return 0, 0, 0, 0
	}
	idx := func(p float64) int {
		k := int(p*float64(n-1) + 0.5)
		if k < 0 {
			k = 0
		}
		if k >= n {
			k = n - 1
		}
		return k
	}
	p50 = sorted[idx(0.50)]
	p95 = sorted[idx(0.95)]
	p99 = sorted[idx(0.99)]
	pmax = sorted[n-1]
	return
}
