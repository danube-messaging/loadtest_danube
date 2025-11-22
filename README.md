# LoadTest_Danube

Danube Messaging load testing tool written in Go. It can spin up multiple producers and consumers across topics, subscription types, partitions, and dispatch strategies, then print a concise console summary with throughput, errors, and end-to-end latency percentiles.

## Prerequisites

- Go 1.24+
- A running Danube broker (single node or cluster)
  - Docs: <https://danube-docs.dev-state.com/>
- macOS/Linux shell

## Install & Build

```bash
go mod tidy
go build -o bin/loadtest ./cmd/loadtest
```

Check the CLI:

```bash
./bin/loadtest --help
./bin/loadtest init --template simple  # writes configs/simple_test.yaml
./bin/loadtest init --template stress  # writes configs/stress_test.yaml
./bin/loadtest init --template patterns  # writes configs/patterns_test.yaml
```

## Quick Start

1) Ensure Danube broker is running and reachable (default used here is 127.0.0.1:6650).

2) Validate a config:

```bash
./bin/loadtest validate --config configs/simple_test.yaml
```

3) Run the load test:

```bash
./bin/loadtest run --config configs/simple_test.yaml
```

You’ll see periodic “Stats:” lines and a final summary including sent/received/errors, tx/rx throughput, and latency percentiles.

## Configuration Overview (YAML)

Top-level keys:

- `test_name`, `description`
- `danube.service_url`: e.g. "127.0.0.1:6650"
- `execution.duration`: test run duration (e.g. "2m")
- `topics[]`: topic definitions (schema, partitions, dispatch)
- `producers[]`: producer groups (topic, count, rate)
- `consumers[]`: consumer groups (topic, subscription, type, count)
- `metrics`: console reporting options (interval)

## Example

Example to test the main Danube usage patterns:

```bash
./bin/loadtest run --config configs/patterns_test.yaml

Loaded test 'patterns_test' targeting 127.0.0.1:6650
Topics: 5, Producers: 5 groups, Consumers: 8 groups
2025/11/22 08:30:11 runner.go:52: Load test started for 2m0s...
2025/11/22 08:30:16 runner.go:64: Stats: elapsed=6s sent=562 recv=893 err=0 tx_mps=93.6 rx_mps=148.8 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=4.0 n=893
2025/11/22 08:30:21 runner.go:64: Stats: elapsed=11s sent=1162 recv=1843 err=0 tx_mps=105.6 rx_mps=167.5 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=4.0 n=1843
2025/11/22 08:30:26 runner.go:64: Stats: elapsed=16s sent=1762 recv=2793 err=0 tx_mps=110.1 rx_mps=174.5 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=12.0 n=2793
2025/11/22 08:30:31 runner.go:64: Stats: elapsed=21s sent=2362 recv=3743 err=0 tx_mps=112.5 rx_mps=178.2 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=12.0 n=3743
2025/11/22 08:30:36 runner.go:64: Stats: elapsed=26s sent=2962 recv=4693 err=0 tx_mps=113.9 rx_mps=180.5 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=12.0 n=4693
2025/11/22 08:30:41 runner.go:64: Stats: elapsed=31s sent=3562 recv=5643 err=0 tx_mps=114.9 rx_mps=182.0 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=12.0 n=5643
2025/11/22 08:30:46 runner.go:64: Stats: elapsed=36s sent=4162 recv=6593 err=0 tx_mps=115.6 rx_mps=183.1 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=12.0 n=6593
2025/11/22 08:30:51 runner.go:64: Stats: elapsed=41s sent=4762 recv=7543 err=0 tx_mps=116.1 rx_mps=184.0 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=12.0 n=7543
2025/11/22 08:30:56 runner.go:64: Stats: elapsed=46s sent=5362 recv=8493 err=0 tx_mps=116.6 rx_mps=184.6 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=12.0 n=8493
2025/11/22 08:31:01 runner.go:64: Stats: elapsed=51s sent=5962 recv=9443 err=0 tx_mps=116.9 rx_mps=185.2 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=12.0 n=9443
2025/11/22 08:31:06 runner.go:64: Stats: elapsed=56s sent=6562 recv=10393 err=0 tx_mps=117.2 rx_mps=185.6 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=12.0 n=10393
2025/11/22 08:31:11 runner.go:64: Stats: elapsed=61s sent=7162 recv=11343 err=0 tx_mps=117.4 rx_mps=185.9 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=12.0 n=11343
2025/11/22 08:31:16 runner.go:64: Stats: elapsed=66s sent=7762 recv=12293 err=0 tx_mps=117.6 rx_mps=186.3 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=22.0 n=12293
2025/11/22 08:31:21 runner.go:64: Stats: elapsed=71s sent=8362 recv=13243 err=0 tx_mps=117.8 rx_mps=186.5 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=22.0 n=13243
2025/11/22 08:31:26 runner.go:64: Stats: elapsed=76s sent=8962 recv=14193 err=0 tx_mps=117.9 rx_mps=186.7 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=22.0 n=14193
2025/11/22 08:31:31 runner.go:64: Stats: elapsed=81s sent=9562 recv=15143 err=0 tx_mps=118.0 rx_mps=186.9 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=22.0 n=15143
2025/11/22 08:31:36 runner.go:64: Stats: elapsed=86s sent=10162 recv=16093 err=0 tx_mps=118.2 rx_mps=187.1 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=22.0 n=16093
2025/11/22 08:31:41 runner.go:64: Stats: elapsed=91s sent=10762 recv=17043 err=0 tx_mps=118.3 rx_mps=187.3 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=22.0 n=17043
2025/11/22 08:31:46 runner.go:64: Stats: elapsed=96s sent=11362 recv=17993 err=0 tx_mps=118.4 rx_mps=187.4 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=22.0 n=17993
2025/11/22 08:31:51 runner.go:64: Stats: elapsed=101s sent=11962 recv=18943 err=0 tx_mps=118.4 rx_mps=187.6 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=22.0 n=18943
2025/11/22 08:31:56 runner.go:64: Stats: elapsed=106s sent=12562 recv=19893 err=0 tx_mps=118.5 rx_mps=187.7 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=22.0 n=19893
2025/11/22 08:32:01 runner.go:64: Stats: elapsed=111s sent=13162 recv=20843 err=0 tx_mps=118.6 rx_mps=187.8 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=22.0 n=20843
2025/11/22 08:32:06 runner.go:64: Stats: elapsed=116s sent=13762 recv=21793 err=0 tx_mps=118.6 rx_mps=187.9 lat(ms): p50=1.0 p95=2.0 p99=2.0 max=22.0 n=21793

2025/11/22 08:32:10 report.go:18: 
===== Load Test Summary =====
2025/11/22 08:32:10 report.go:19: Test:        patterns_test
2025/11/22 08:32:10 report.go:20: Broker:      127.0.0.1:6650
2025/11/22 08:32:10 report.go:21: Duration:    2m0s (elapsed 120.0s)
2025/11/22 08:32:10 report.go:22: Messages:    sent=14242  received=22553  errors=0
2025/11/22 08:32:10 report.go:23: Throughput:  tx=118.7 msg/s  rx=187.9 msg/s
2025/11/22 08:32:10 report.go:25: Latency(ms): p50=1.0  p95=2.0  p99=2.0  max=22.0  samples=22553
2025/11/22 08:32:10 report.go:29: Integrity:   loss=0  duplicates=0
2025/11/22 08:32:10 report.go:41: SLA:         keys_in_sla=11  keys_out_sla=0  total_keys=11
2025/11/22 08:32:10 report.go:55: Worst keys (top 5):
2025/11/22 08:32:10 report.go:65: ==============================
```

## What Metrics Are Reported

- Messages sent, received, errors
- Throughput (tx/rx msgs/sec)
- End-to-end latency (ms): p50, p95, p99, max, sample count (computed from message PublishTime)
- Integrity: estimated message loss and duplicate counts per (topic, subscription, producer)

### Results Export (optional)

- Set `metrics.export_path: "./results"` in your config to also write a JSON file with the final snapshot and a summary of the config.
- The export includes integrity breakdown entries per (topic, subscription, producer) and stable JSON keys for easy diffing.
