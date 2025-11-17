package config

// Example configuration templates exposed via `loadtest init`.

const SimpleTemplate = `# Test configuration
test_name: "simple_throughput_test"
description: "Basic throughput test with one topic and shared subscription"

danube:
  service_url: "127.0.0.1:6650"

execution:
  duration: "2m"
  warmup_duration: "5s"
  cooldown_duration: "3s"

topics:
  - name: "/default/test"
    partitions: 0
    schema_type: "string"

producers:
  - name: "test_producers"
    topic: "/default/test"
    count: 3
    rate_per_second: 50
    message_size: 256

consumers:
  - name: "test_consumers"
    topic: "/default/test"
    subscription: "sub_shared"
    subscription_type: "shared"
    count: 3

metrics:
  enabled: true
  report_interval: "5s"
  percentiles: [50, 95, 99]
  output_format: "terminal"
  export_path: "./results"
  collect:
    - producer_throughput
    - consumer_throughput
    - end_to_end_latency
    - producer_latency
    - error_rates
`

const StressTemplate = `# Stress configuration across multiple topics
test_name: "multi_topic_stress_test"
description: "Stress test with multiple topics, partitions and subscription types"

danube:
  service_url: "127.0.0.1:6650"

execution:
  duration: "5m"
  warmup_duration: "10s"
  cooldown_duration: "5s"

topics:
  - name: "/default/orders"
    partitions: 3
    schema_type: "json"
    json_schema: |
      {"type":"object","properties":{"order_id":{"type":"string"},"amount":{"type":"number"},"ts":{"type":"integer"}},"required":["order_id","amount"]}
  - name: "/default/events"
    partitions: 1
    schema_type: "string"
  - name: "/default/metrics"
    partitions: 0
    schema_type: "int64"

producers:
  - name: "order_producers"
    topic: "/default/orders"
    count: 5
    rate_per_second: 100
    message_size: 1024
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

consumers:
  - name: "order_processor"
    topic: "/default/orders"
    subscription: "order_processing"
    subscription_type: "shared"
    count: 3
  - name: "order_analytics"
    topic: "/default/orders"
    subscription: "analytics"
    subscription_type: "exclusive"
    count: 1
  - name: "order_backup"
    topic: "/default/orders"
    subscription: "backup"
    subscription_type: "failover"
    count: 2
  - name: "event_handler"
    topic: "/default/events"
    subscription: "event_processing"
    subscription_type: "shared"
    count: 2

metrics:
  enabled: true
  report_interval: "5s"
  percentiles: [50, 95, 99, 99.9]
  output_format: "terminal"
  export_path: "./results"
  collect:
    - producer_throughput
    - consumer_throughput
    - end_to_end_latency
    - producer_latency
    - message_loss
    - error_rates
`
