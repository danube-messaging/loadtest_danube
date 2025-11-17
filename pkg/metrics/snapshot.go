package metrics

import (
	"encoding/json"
)

// Snapshot is an immutable snapshot of collected metrics used for reporting/export.
type Snapshot struct {
	ElapsedSec         float64          `json:"elapsed_sec"`
	MessagesSent       uint64           `json:"messages_sent"`
	MessagesReceived   uint64           `json:"messages_received"`
	Errors             uint64           `json:"errors"`
	ThroughputSent     float64          `json:"throughput_sent"`
	ThroughputRecv     float64          `json:"throughput_recv"`
	LatencyP50Ms       float64          `json:"latency_p50_ms"`
	LatencyP95Ms       float64          `json:"latency_p95_ms"`
	LatencyP99Ms       float64          `json:"latency_p99_ms"`
	LatencyMaxMs       float64          `json:"latency_max_ms"`
	LatencySamples     int              `json:"latency_samples"`
	Duplicates         uint64           `json:"duplicates"`
	EstimatedLoss      uint64           `json:"estimated_loss"`
	IntegrityBreakdown []IntegrityEntry `json:"integrity_breakdown,omitempty"`
}

// IntegrityEntry holds per-key integrity metrics for export and reporting
type IntegrityEntry struct {
	Topic        string `json:"topic"`
	Subscription string `json:"subscription"`
	Producer     string `json:"producer"`
	Min          uint64 `json:"min"`
	Max          uint64 `json:"max"`
	UniqueSeen   uint64 `json:"unique_seen"`
	Loss         uint64 `json:"loss"`
	Duplicates   uint64 `json:"duplicates"`
}

// MarshalSnapshot returns a JSON []byte of the snapshot for export.
func (s Snapshot) MarshalJSONBytes() ([]byte, error) {
	return json.Marshal(s)
}

// Helpers used by Collector.Snapshot
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
