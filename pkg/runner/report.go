package runner

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/danube-messaging/loadtest_danube/pkg/config"
	"github.com/danube-messaging/loadtest_danube/pkg/metrics"
)

// printSummary prints the final human-readable summary including SLA and top-5 worst keys
func printSummary(cfg *config.Config, snap metrics.Snapshot, dur time.Duration) {
	log.Println("\n===== Load Test Summary =====")
	log.Printf("Test:        %s", cfg.TestName)
	log.Printf("Broker:      %s", cfg.Danube.ServiceURL)
	log.Printf("Duration:    %s (elapsed %.1fs)", dur, snap.ElapsedSec)
	log.Printf("Messages:    sent=%d  received=%d  errors=%d", snap.MessagesSent, snap.MessagesReceived, snap.Errors)
	log.Printf("Throughput:  tx=%.1f msg/s  rx=%.1f msg/s", snap.ThroughputSent, snap.ThroughputRecv)
	if snap.LatencySamples > 0 {
		log.Printf("Latency(ms): p50=%.1f  p95=%.1f  p99=%.1f  max=%.1f  samples=%d", snap.LatencyP50Ms, snap.LatencyP95Ms, snap.LatencyP99Ms, snap.LatencyMaxMs, snap.LatencySamples)
	} else {
		log.Printf("Latency(ms): no samples (enable string/json payloads to measure)")
	}
	log.Printf("Integrity:   loss=%d  duplicates=%d", snap.EstimatedLoss, snap.Duplicates)

	// SLA and top 5 worst keys
	if len(snap.IntegrityBreakdown) > 0 {
		total := len(snap.IntegrityBreakdown)
		inSLA := 0
		for _, e := range snap.IntegrityBreakdown {
			if e.Loss == 0 && e.Duplicates == 0 {
				inSLA++
			}
		}
		outSLA := total - inSLA
		log.Printf("SLA:         keys_in_sla=%d  keys_out_sla=%d  total_keys=%d", inSLA, outSLA, total)
		// Sort worst first: by loss desc, then duplicates desc
		bd := append([]metrics.IntegrityEntry(nil), snap.IntegrityBreakdown...)
		sort.Slice(bd, func(i, j int) bool {
			if bd[i].Loss == bd[j].Loss {
				return bd[i].Duplicates > bd[j].Duplicates
			}
			return bd[i].Loss > bd[j].Loss
		})
		maxShow := 5
		if len(bd) < maxShow {
			maxShow = len(bd)
		}
		if maxShow > 0 {
			log.Printf("Worst keys (top %d):", maxShow)
			for k := 0; k < maxShow; k++ {
				e := bd[k]
				if e.Loss == 0 && e.Duplicates == 0 {
					break
				}
				log.Printf("  - %s | %s | %s : loss=%d dup=%d range=[%d..%d] seen=%d", e.Topic, e.Subscription, e.Producer, e.Loss, e.Duplicates, e.Min, e.Max, e.UniqueSeen)
			}
		}
	}
	log.Println("==============================\n")
}

// exportResults writes a JSON file with snapshot and run description if ExportPath is configured
func exportResults(cfg *config.Config, snap metrics.Snapshot) {
	if cfg.Metrics.ExportPath == "" {
		return
	}
	if err := os.MkdirAll(cfg.Metrics.ExportPath, 0o755); err != nil {
		log.Printf("failed to create export dir: %v", err)
		return
	}
	ts := time.Now().Format("20060102_150405")
	out := struct {
		TestName    string           `json:"test_name"`
		Description string           `json:"description,omitempty"`
		ServiceURL  string           `json:"service_url"`
		DurationSec float64          `json:"duration_sec"`
		Snapshot    metrics.Snapshot `json:"snapshot"`
		Config      struct {
			Producers int `json:"producers"`
			Consumers int `json:"consumers"`
			Topics    int `json:"topics"`
		} `json:"config_summary"`
	}{
		TestName:    cfg.TestName,
		Description: cfg.Description,
		ServiceURL:  cfg.Danube.ServiceURL,
		DurationSec: snap.ElapsedSec,
		Snapshot:    snap,
		Config: struct {
			Producers int `json:"producers"`
			Consumers int `json:"consumers"`
			Topics    int `json:"topics"`
		}{Producers: len(cfg.Producers), Consumers: len(cfg.Consumers), Topics: len(cfg.Topics)},
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		log.Printf("export marshal error: %v", err)
		return
	}
	path := filepath.Join(cfg.Metrics.ExportPath, fmt.Sprintf("%s_%s.json", cfg.TestName, ts))
	if err := os.WriteFile(path, b, 0o644); err != nil {
		log.Printf("export write error: %v", err)
	} else {
		log.Printf("Results exported to %s", path)
	}
}
