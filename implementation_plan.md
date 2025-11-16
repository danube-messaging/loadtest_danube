# Danube Load Testing Tool - Implementation Plan

## Overview
Transform the LoadTest_Danube repository into a comprehensive load testing framework for Danube messaging that supports concurrent producers/consumers, multiple topics, various subscription types, and detailed metrics collection.

---

## Current State Analysis

### ✅ What We Have
- Basic producer/consumer with HTTP API
- YAML-based configuration
- Support for multiple schema types (JSON, String, Number)
- Integration with danube-go SDK

### ❌ What's Missing
- No metrics collection or reporting
- No concurrent load generation
- Limited subscription type testing
- No end-to-end latency measurement
- No scalability for high-throughput testing
- Manual test orchestration

---

## Proposed Architecture

### Directory Structure
```
LoadTest_Danube/
├── cmd/
│   └── loadtest/
│       └── main.go                    # CLI entry point
├── pkg/
│   ├── config/
│   │   ├── config.go                  # Configuration structures
│   │   ├── parser.go                  # YAML parsing
│   │   └── validator.go               # Config validation
│   ├── producer/
│   │   ├── pool.go                    # Producer pool management
│   │   ├── worker.go                  # Individual producer worker
│   │   └── rate_limiter.go            # Rate limiting logic
│   ├── consumer/
│   │   ├── pool.go                    # Consumer pool management
│   │   ├── worker.go                  # Individual consumer worker
│   │   └── tracker.go                 # Message tracking
│   ├── metrics/
│   │   ├── collector.go               # Metrics collection
│   │   ├── reporter.go                # Real-time & final reporting
│   │   ├── types.go                   # Metric types & aggregations
│   │   └── dashboard.go               # Terminal UI dashboard
│   ├── workload/
│   │   ├── scenario.go                # Test scenario orchestration
│   │   ├── generator.go               # Message payload generation
│   │   └── patterns.go                # Load patterns (constant, burst, ramp)
│   └── utils/
│       ├── logger.go                  # Structured logging
│       └── signals.go                 # Graceful shutdown
├── configs/
│   ├── simple_test.yaml               # Basic test scenario
│   ├── stress_test.yaml               # High-load scenario
│   ├── subscription_test.yaml         # Test all subscription types
│   ├── partition_test.yaml            # Multi-partition testing
│   └── failover_test.yaml             # Failover scenario
├── scripts/
│   ├── run_basic_test.sh
│   ├── run_stress_test.sh
│   └── setup_topics.sh
├── results/                           # Test results output
├── go.mod
├── go.sum
├── README.md
└── implementation_plan.md             # This file
```

---

## Configuration Schema

### Example Configuration File
```yaml
# Test configuration
test_name: "multi_topic_stress_test"
description: "Stress test with multiple topics and subscription types"

# Danube broker connection
danube:
  service_url: "127.0.0.1:6650"
  connection_timeout: "10s"

# Test execution parameters
execution:
  duration: "5m"                       # Test duration
  warmup_duration: "10s"               # Warmup period before metrics
  cooldown_duration: "5s"              # Cooldown after stopping producers

# Topic definitions
topics:
  - name: "/default/orders"
    partitions: 3
    schema_type: "json"
    json_schema: |
      {
        "type": "object",
        "properties": {
          "order_id": {"type": "string"},
          "customer_id": {"type": "string"},
          "amount": {"type": "number"},
          "timestamp": {"type": "integer"}
        },
        "required": ["order_id", "customer_id", "amount"]
      }
  
  - name: "/default/events"
    partitions: 1
    schema_type: "string"
  
  - name: "/default/metrics"
    partitions: 0                      # Non-partitioned
    schema_type: "int64"

# Producer configurations
producers:
  - name: "order_producers"
    topic: "/default/orders"
    count: 5                           # 5 concurrent producers
    rate_per_second: 100              # messages/sec per producer
    message_size: 1024                # bytes (for payload generation)
    batch_size: 10                    # optional batching
    
  - name: "event_producers"
    topic: "/default/events"
    count: 3
    rate_per_second: 50
    message_size: 512
    
  - name: "metric_producers"
    topic: "/default/metrics"
    count: 2
    rate_per_second: 200
    message_size: 64

# Consumer configurations
consumers:
  - name: "order_processor"
    topic: "/default/orders"
    subscription: "order_processing"
    subscription_type: "shared"       # shared, exclusive, failover
    count: 3                          # 3 consumers sharing load
    ack_timeout: "30s"
    
  - name: "order_analytics"
    topic: "/default/orders"
    subscription: "analytics"
    subscription_type: "exclusive"
    count: 1
  
  - name: "order_backup"
    topic: "/default/orders"
    subscription: "backup"
    subscription_type: "failover"
    count: 2                          # 1 active, 1 standby
    
  - name: "event_handler"
    topic: "/default/events"
    subscription: "event_processing"
    subscription_type: "shared"
    count: 2

# Metrics and reporting
metrics:
  enabled: true
  report_interval: "5s"               # Real-time reporting interval
  percentiles: [50, 95, 99, 99.9]    # Latency percentiles
  output_format: "terminal"           # terminal, json, prometheus
  export_path: "./results"            # Results export directory
  
  collect:
    - producer_throughput
    - consumer_throughput
    - end_to_end_latency
    - producer_latency
    - message_loss
    - error_rates
```

---

## Metrics to Collect

### Producer Metrics
| Metric | Description | Type |
|--------|-------------|------|
| `messages_sent_total` | Total messages sent | Counter |
| `messages_sent_per_topic` | Messages per topic | Counter |
| `bytes_sent_total` | Total bytes sent | Counter |
| `send_latency_ms` | Time to send message | Histogram |
| `send_errors_total` | Send failures | Counter |
| `producer_throughput_mps` | Messages per second | Gauge |
| `producer_throughput_bps` | Bytes per second | Gauge |
| `active_producers` | Number of active producers | Gauge |

### Consumer Metrics
| Metric | Description | Type |
|--------|-------------|------|
| `messages_received_total` | Total messages received | Counter |
| `messages_received_per_topic` | Messages per topic | Counter |
| `bytes_received_total` | Total bytes received | Counter |
| `receive_latency_ms` | Time to receive message | Histogram |
| `end_to_end_latency_ms` | Send to receive time | Histogram |
| `ack_latency_ms` | Time to acknowledge | Histogram |
| `ack_errors_total` | Acknowledgment failures | Counter |
| `consumer_throughput_mps` | Messages per second | Gauge |
| `consumer_throughput_bps` | Bytes per second | Gauge |
| `active_consumers` | Number of active consumers | Gauge |
| `messages_lost` | Detected message loss | Counter |
| `messages_duplicated` | Detected duplicates | Counter |

### System Metrics
- Active connections
- Memory usage
- Goroutine count
- Test duration
- Error breakdown by type

---

## Test Scenarios

### Scenario A: Basic Throughput Test
**Objective:** Measure baseline throughput and latency

**Configuration:**
- 1 topic (non-partitioned)
- 3 producers @ 100 msg/sec each
- 3 consumers (shared subscription)
- Duration: 5 minutes

**Expected Outcomes:**
- Stable throughput ~300 msg/sec
- Low latency (< 10ms p99)
- Zero message loss

---

### Scenario B: Subscription Type Validation
**Objective:** Verify message distribution across subscription types

**Configuration:**
- 1 topic
- 1 producer @ 100 msg/sec
- 3 subscriptions:
  - Exclusive (1 consumer)
  - Shared (3 consumers)
  - Failover (2 consumers)

**Expected Outcomes:**
- Exclusive: 1 consumer gets all messages
- Shared: Messages distributed round-robin
- Failover: Only 1 consumer active

---

### Scenario C: Multi-Topic Stress Test
**Objective:** Test high load across multiple topics

**Configuration:**
- 5-10 topics (mixed partitioned/non-partitioned)
- 20+ producers @ varying rates
- 30+ consumers across subscriptions
- Duration: 10 minutes
- Target: 5000+ msg/sec aggregate

**Expected Outcomes:**
- System handles high load
- Latency remains acceptable
- No message loss
- Resource usage tracking

---

### Scenario D: Partition Load Distribution
**Objective:** Test partition distribution

**Configuration:**
- 1 partitioned topic (5 partitions)
- 1 producer for the topic (producers operate at topic level; partitions are internal)
- 5 consumers (shared subscription)
- Duration: 5 minutes

**Expected Outcomes:**
- Even distribution across partitions (observed via broker behavior)
- Consumers balanced across partitions
- Consistent throughput per partition

---

### Scenario E: Failover Test
**Objective:** Validate failover behavior

**Configuration:**
- 1 topic
- 1 producer @ 100 msg/sec
- Failover subscription with 3 consumers
- Simulate consumer failures mid-test

**Expected Outcomes:**
- Automatic failover to standby consumer
- No message loss during failover
- Minimal latency spike

---

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1)
**Duration:** 3-4 days

**Tasks:**
- [x] Set up new project structure
- [x] Design configuration schema
- [x] Implement YAML parsing and validation
- [x] Create CLI framework (Cobra)
- [x] Implement graceful shutdown handling
- [ ] Set up structured logging

**Deliverables:**
- CLI skeleton with `run`, `validate`, `init` commands
- Configuration loader with validation
- Basic project structure
- Initial example configs (simple, stress)

**Dependencies:**
```go
github.com/spf13/cobra              // CLI framework
github.com/spf13/viper              // Configuration
go.uber.org/zap                     // Logging
gopkg.in/yaml.v3                    // YAML parsing
```

---

### Phase 2: Metrics System (Week 1-2)
**Duration:** 2-3 days

**Tasks:**
- [ ] Define metric types and structures
- [ ] Implement thread-safe metric collectors
- [ ] Build aggregation logic (percentiles, rates)
- [ ] Create real-time reporter
- [ ] Implement terminal dashboard (bubbletea)
- [ ] Add JSON export functionality

**Deliverables:**
- Metrics collection framework (MVP collector added)
- Console summary printed at the end of each run (no external systems)
- Local summary file export in results/ (optional JSON/CSV for later viewing)

**Dependencies:**
```go
github.com/charmbracelet/bubbletea  // Terminal UI
github.com/charmbracelet/lipgloss   // Terminal styling
sync/atomic                          // Atomic operations
```

---

### Phase 3: Producer Implementation (Week 2)
**Status:** In progress (MVP implemented: concurrent workers per topic with rate limiting; basic error and throughput metrics). Next: explicit schema configuration and reliable dispatch option.
**Duration:** 3-4 days

**Tasks:**
- [ ] Create producer pool manager
- [ ] Implement producer workers with goroutines
- [ ] Add rate limiting per producer
- [ ] Build message payload generator
- [ ] Implement producer metrics tracking
- [ ] Add support for all schema types
- [ ] Handle producer lifecycle (create, run, stop)
- [ ] Add batching support (optional)

**Deliverables:**
- Concurrent producer pool (MVP implemented)
- Rate-limited message generation (per-worker limiter)
- Producer metrics integration (messages sent/errors)
- Topic-level producers only (partitions are internal and managed by Danube)

**Dependencies:**
```go
golang.org/x/time/rate              // Rate limiting
github.com/danube-messaging/danube-go
```

---

### Phase 4: Consumer Implementation (Week 2)
**Status:** In progress (MVP implemented: consumer workers per subscription, ack handling, basic throughput/error metrics). Next: end-to-end latency and subscription test matrix.
**Duration:** 3-4 days

**Tasks:**
- [ ] Create consumer pool manager
- [ ] Implement consumer workers
- [ ] Add message acknowledgment handling
- [ ] Implement end-to-end latency tracking
- [ ] Build message tracking (detect loss/duplication)
- [ ] Add consumer metrics tracking
- [ ] Support all subscription types
- [ ] Handle consumer lifecycle

**Deliverables:**
- Concurrent consumer pool (MVP implemented)
- Message tracking system (planned)
- Consumer metrics integration (messages received/errors)

---

### Phase 5: Scenario Orchestration (Week 3)
**Duration:** 2-3 days

**Tasks:**
- [ ] Implement test scenario runner
- [ ] Add warmup/cooldown logic
- [ ] Coordinate producer/consumer lifecycle
- [ ] Implement duration-based test control
- [ ] Add progress tracking
- [ ] Create scenario templates

**Deliverables:**
- Complete test orchestration
- 5+ example scenario configs
- Test runner with progress tracking

---

### Phase 6: Reporting & Analysis (Week 3)
**Duration:** 2 days

**Tasks:**
- [ ] Implement final report generation
- [ ] Add summary statistics
- [ ] Create comparison tool
- [ ] Add result visualization helpers
- [ ] Export to multiple formats (JSON, CSV)

**Deliverables:**
- Comprehensive final reports
- Result comparison tool
- Multiple export formats

---

 

 

## Technology Stack

### Core Dependencies
| Package | Purpose | Priority |
|---------|---------|----------|
| `github.com/danube-messaging/danube-go` | Danube SDK | Required |
| `github.com/spf13/cobra` | CLI framework | High |
| `gopkg.in/yaml.v3` | YAML parsing | High |
| `golang.org/x/time/rate` | Rate limiting | High |

### Reporting
Console summary after each run (printed by the runner). Optional local export to results/ as JSON or CSV for offline viewing.

### Testing
| Package | Purpose | Priority |
|---------|---------|----------|
| `github.com/stretchr/testify` | Testing utilities | High |
| `github.com/golang/mock` | Mocking | Medium |

---

## CLI Interface Design

### Command Structure
```bash
loadtest [command] [flags]
```

### Commands

#### `run` - Execute a test scenario
```bash
# Run with config file
loadtest run --config configs/stress_test.yaml

# Override config parameters
loadtest run --config configs/basic.yaml \
  --duration 10m \
  --producers 10 \
  --consumers 15

# Quick test with defaults
loadtest run --quick \
  --topic /default/test \
  --producers 5 \
  --consumers 5 \
  --duration 1m
```

#### `validate` - Validate configuration
```bash
loadtest validate --config configs/stress_test.yaml
```

#### `init` - Generate example configs
```bash
# Generate all templates
loadtest init

# Generate specific template
loadtest init --template stress
loadtest init --template subscription
```

#### `results` - View test results
```bash
# View specific result
loadtest results --path ./results/test_20240116_154530

# Compare results
loadtest results --compare ./results/test1 ./results/test2

# List all results
loadtest results --list
```

#### `version` - Show version
```bash
loadtest version
```

### Global Flags
- `--verbose, -v` - Verbose output
- `--debug` - Debug logging
- `--quiet, -q` - Minimal output
- `--no-color` - Disable colored output

---

## Key Design Decisions

### 1. Concurrency Model
- **Producer Pool:** Goroutine per producer worker
- **Consumer Pool:** Goroutine per consumer worker
- **Rate Limiting:** Token bucket per producer
- **Metrics:** Lock-free atomic operations where possible

### 2. Message Tracking
- Embed sequence numbers in message payloads
- Track sent messages in ring buffer (for loss detection)
- Compare sent vs received for accuracy

### 3. Error Handling
- Retry logic for transient failures
- Configurable error thresholds
- Fail-fast vs. continue-on-error modes

### 4. Resource Management
- Graceful shutdown on SIGINT/SIGTERM
- Connection pooling for efficiency
- Memory-bounded buffers

### 5. Extensibility
- Plugin system for custom metrics
- Configurable payload generators
- Modular architecture for easy enhancement

---

## Success Criteria

### Functional Requirements
- ✅ Support 1000+ concurrent producers/consumers
- ✅ Handle 10,000+ messages/second aggregate
- ✅ All subscription types working correctly
- ✅ Accurate metrics collection (< 1% error)
- ✅ Zero message loss in stable conditions

### Performance Requirements
- ✅ Minimal overhead (< 5% CPU for metrics)
- ✅ Low memory footprint (< 100MB baseline)
- ✅ Real-time dashboard updates (< 1s lag)

### Usability Requirements
- ✅ Simple YAML configuration
- ✅ Clear CLI interface
- ✅ Informative error messages
- ✅ Comprehensive documentation

---

## Risk Mitigation

### Risk: SDK Limitations
**Mitigation:** Review danube-go SDK capabilities early; implement workarounds if needed

### Risk: Performance Bottlenecks
**Mitigation:** Profile early and often; optimize hot paths; use buffered channels

### Risk: Metric Overhead
**Mitigation:** Use lock-free operations; sample metrics if needed; benchmark impact

### Risk: Complex Configuration
**Mitigation:** Provide validation; create templates; document examples

---

## Future Enhancements (Post-Launch)

1. **Distributed Testing**
   - Coordinate multiple load test instances
   - Aggregate metrics across instances

2. **Advanced Analytics**
   - Historical trending
   - Anomaly detection
   - Performance regression detection

3. **Cloud Integration**
   - Kubernetes operator
   - Cloud-native deployment
   - Auto-scaling test load

4. **Web UI**
   - Real-time web dashboard
   - Historical results browser
   - Test management interface

5. **CI/CD Integration**
   - GitHub Actions integration
   - Performance gate enforcement
   - Automated benchmarking

---

## Timeline Summary

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| Phase 1: Core Infrastructure | 3-4 days | CLI, Config, Structure |
| Phase 2: Metrics System | 2-3 days | Metrics, Dashboard |
| Phase 3: Producer Implementation | 3-4 days | Producer pools |
| Phase 4: Consumer Implementation | 3-4 days | Consumer pools |
| Phase 5: Scenario Orchestration | 2-3 days | Test runner |
| Phase 6: Reporting & Analysis | 2 days | Reports, exports |
| Phase 7: Advanced Features | 3-5 days | Optional features |
| Phase 8: Testing & Documentation | 2-3 days | Tests, docs |
| **Total** | **~3-4 weeks** | Full implementation |

**MVP:** 1-2 days for basic functionality (in progress)

---

## Getting Started

### Immediate Next Steps
1. Review and approve this plan
2. Choose between MVP or full implementation
3. Set up development environment
4. Begin Phase 1 implementation

### Development Environment Setup
```bash
# Clone repository
cd LoadTest_Danube

# Initialize Go modules
go mod tidy

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run Danube broker locally (Docker)
docker run -p 6650:6650 danube/danube:latest
```

---

## Questions for Clarification

1. **Priority:** MVP first or full implementation?
2. **Timeline:** Any specific deadlines?
3. **Features:** Any additional features not covered?
4. **Deployment:** Local only or cloud deployment needed?
5. **Integration:** Need integration with existing monitoring systems?

---

*Last Updated: 2025-11-16*
*Status: In Progress*
*Author: Cascade AI*
