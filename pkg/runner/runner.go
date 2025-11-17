package runner

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/danube-messaging/loadtest_danube/pkg/config"
	"github.com/danube-messaging/loadtest_danube/pkg/consumer"
	"github.com/danube-messaging/loadtest_danube/pkg/metrics"
	"github.com/danube-messaging/loadtest_danube/pkg/producer"
	"github.com/danube-messaging/loadtest_danube/pkg/utils"
)

// Run executes the scenario described by cfg for the specified duration.
func Run(cfg *config.Config) error {
	// Context with interrupt and duration
	base := context.Background()
	ctx := utils.WithInterrupt(base)
	dur, err := time.ParseDuration(cfg.Execution.Duration)
	if err != nil {
		return fmt.Errorf("invalid execution.duration: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, dur)
	defer cancel()

	m := metrics.NewCollector()

	// Start pools
	var wg sync.WaitGroup
	prodPool := producer.NewPool(cfg.Danube.ServiceURL, cfg, m)
	consPool := consumer.NewPool(cfg.Danube.ServiceURL, cfg, m)

	// Start producers first to ensure topics are created on the broker
	prodPool.Start(ctx, &wg)
	// Small delay to avoid races where consumers subscribe before topics exist
	time.Sleep(1 * time.Second)
	consPool.Start(ctx, &wg)

	// periodic reporting
	reportEvery := 5 * time.Second
	if cfg.Metrics.ReportInterval != "" {
		if d, err := time.ParseDuration(cfg.Metrics.ReportInterval); err == nil {
			reportEvery = d
		}
	}
	ticker := time.NewTicker(reportEvery)
	defer ticker.Stop()

	log.Printf("Load test started for %s...", dur)
	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			snap := m.Snapshot()
			// Pretty summary and optional export delegated to helpers
			printSummary(cfg, snap, dur)
			exportResults(cfg, snap)
			return nil
		case <-ticker.C:
			snap := m.Snapshot()
			log.Printf("Stats: elapsed=%.0fs sent=%d recv=%d err=%d tx_mps=%.1f rx_mps=%.1f lat(ms): p50=%.1f p95=%.1f p99=%.1f max=%.1f n=%d",
				snap.ElapsedSec, snap.MessagesSent, snap.MessagesReceived, snap.Errors, snap.ThroughputSent, snap.ThroughputRecv,
				snap.LatencyP50Ms, snap.LatencyP95Ms, snap.LatencyP99Ms, snap.LatencyMaxMs, snap.LatencySamples)
		}
	}
}
