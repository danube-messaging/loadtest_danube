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

	// message tracking per topic+subscription+producer
	trackers map[trackerKey]*seqTracker
}

type trackerKey struct {
	Topic        string
	Subscription string
	Producer     string
}

type seqTracker struct {
	Min        uint64
	Max        uint64
	Seen       map[uint64]struct{}
	Duplicates uint64
}

func NewCollector() *Collector {
	return &Collector{Start: time.Now(), trackers: make(map[trackerKey]*seqTracker)}
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

// RecordSeq records observed sequence for a given topic+subscription+producer.
func (c *Collector) RecordSeq(topic, subscription, producer string, seq uint64) {
	k := trackerKey{Topic: topic, Subscription: subscription, Producer: producer}
	c.mu.Lock()
	t, ok := c.trackers[k]
	if !ok {
		t = &seqTracker{Min: seq, Max: seq, Seen: make(map[uint64]struct{})}
		c.trackers[k] = t
	}
	if seq < t.Min {
		t.Min = seq
	}
	if seq > t.Max {
		t.Max = seq
	}
	if _, exists := t.Seen[seq]; exists {
		t.Duplicates++
	} else {
		t.Seen[seq] = struct{}{}
	}
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
	// compute duplicate and loss estimates and build breakdown per key
	var dup uint64
	var loss uint64
	var breakdown []IntegrityEntry
	for k, t := range c.trackers {
		entry := IntegrityEntry{
			Topic:        k.Topic,
			Subscription: k.Subscription,
			Producer:     k.Producer,
			Min:          t.Min,
			Max:          t.Max,
			UniqueSeen:   uint64(len(t.Seen)),
			Duplicates:   t.Duplicates,
		}
		if t.Max >= t.Min {
			expected := (t.Max - t.Min + 1)
			if expected > entry.UniqueSeen {
				entry.Loss = expected - entry.UniqueSeen
			}
		}
		dup += entry.Duplicates
		loss += entry.Loss
		breakdown = append(breakdown, entry)
	}
	c.mu.Unlock()
	sort.Float64s(lcopy)
	p50, p95, p99, pmax := percentiles(lcopy)
	return Snapshot{
		ElapsedSec:         elapsed,
		MessagesSent:       sent,
		MessagesReceived:   recv,
		Errors:             errs,
		ThroughputSent:     rate(sent, elapsed),
		ThroughputRecv:     rate(recv, elapsed),
		LatencyP50Ms:       p50,
		LatencyP95Ms:       p95,
		LatencyP99Ms:       p99,
		LatencyMaxMs:       pmax,
		LatencySamples:     len(lcopy),
		Duplicates:         dup,
		EstimatedLoss:      loss,
		IntegrityBreakdown: breakdown,
	}
}
