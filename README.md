# LoadTest_Danube

Danube Messaging load testing tool written in Go. It can spin up multiple producers and consumers across topics, subscription types, partitions, and dispatch strategies, then print a concise console summary with throughput, errors, and end-to-end latency percentiles.

## Prerequisites
- Go 1.24+
- A running Danube broker (single node or cluster)
  - Docs: https://danube-docs.dev-state.com/
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

Example (excerpt):
```yaml
danube:
  service_url: "127.0.0.1:6650"

execution:
  duration: "2m"

topics:
  - name: "/default/load_simple_test"
    partitions: 0
    schema_type: "string"            # json | string | int64 (number)
    dispatch_strategy: "non_reliable" # reliable | non_reliable
    # json_schema: '{"type":...}'    # required when schema_type=json

producers:
  - name: "test_producers"
    topic: "/default/load_simple_test"
    count: 3                          # concurrent producer workers
    rate_per_second: 100              # per worker
    message_size: 256                 # bytes (approx for string payloads)

consumers:
  - name: "test_consumers"
    topic: "/default/load_simple_test"
    subscription: "sub_shared"
    subscription_type: "shared"      # shared | exclusive | failover
    count: 3                          # concurrent consumers for this sub

metrics:
  enabled: true
  report_interval: "5s"
```

### Topics
- `partitions`: integer >= 0
  - Non-zero uses `WithPartitions(int32(partitions))` on producer.
- `schema_type`:
  - `string`: simple string payloads.
  - `json`: producer sets a JSON schema (must provide `json_schema`).
  - `int64` or `number`: numeric messages.
- `dispatch_strategy`:
  - `reliable`: producer uses `WithDispatchStrategy(danube.NewReliableDispatchStrategy())`.
  - `non_reliable` or omitted: no dispatch strategy set.

### Producers
- Each producer group launches `count` concurrent workers.
- Worker names are unique: `<name>-<index>` to avoid collisions in broker.
- Rate limiting is per-worker using a token bucket.

### Consumers
- Each consumer group launches `count` concurrent workers.
- Worker names are unique: `<name>-<index>`.
- Supported subscription types: `exclusive`, `shared`, `failover`.
- Reliable dispatch requires consumers to acknowledge messages.

## What Metrics Are Reported
- Messages sent, received, errors
- Throughput (tx/rx msgs/sec)
- End-to-end latency (ms): p50, p95, p99, max, sample count (computed from message PublishTime)
- Integrity: estimated message loss and duplicate counts per (topic, subscription, producer)

Notes:
- Latency uses broker/client `PublishTime`; payloads no longer embed timestamps.
- Sequence numbers are attached as message attributes (e.g., `seq` and `producer`) for integrity tracking.

### Results Export (optional)
- Set `metrics.export_path: "./results"` in your config to also write a JSON file with the final snapshot and a summary of the config.
- The export includes integrity breakdown entries per (topic, subscription, producer) and stable JSON keys for easy diffing.

## Common Scenarios

1) Shared subscription distribution
```bash
./bin/loadtest run --config configs/simple_test.yaml
```

2) Reliable dispatch with partitions
```bash
./bin/loadtest run --config configs/stress_test.yaml
```

3) Patterns coverage (mixed schemas, partitions, and subs)
```bash
./bin/loadtest run --config configs/patterns_test.yaml
```


