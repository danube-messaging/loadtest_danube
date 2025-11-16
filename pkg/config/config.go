package config

// Core configuration structures for load testing scenarios.

type Config struct {
	TestName    string `yaml:"test_name"`
	Description string `yaml:"description,omitempty"`

	Danube    DanubeConfig    `yaml:"danube"`
	Execution ExecutionConfig `yaml:"execution"`

	Topics    []Topic         `yaml:"topics"`
	Producers []ProducerGroup `yaml:"producers"`
	Consumers []ConsumerGroup `yaml:"consumers"`

	Metrics MetricsConfig `yaml:"metrics"`
}

type DanubeConfig struct {
	ServiceURL        string `yaml:"service_url"`
	ConnectionTimeout string `yaml:"connection_timeout,omitempty"`
}

type ExecutionConfig struct {
	Duration         string `yaml:"duration"`
	WarmupDuration   string `yaml:"warmup_duration,omitempty"`
	CooldownDuration string `yaml:"cooldown_duration,omitempty"`
}

type Topic struct {
	Name             string `yaml:"name"`
	Partitions       int    `yaml:"partitions"`  // 0 or omitted means non-partitioned
	SchemaType       string `yaml:"schema_type"` // json|string|int64|number
	JSONSchema       string `yaml:"json_schema,omitempty"`
	DispatchStrategy string `yaml:"dispatch_strategy,omitempty"` // reliable|non_reliable (default non_reliable)
}

type ProducerGroup struct {
	Name          string `yaml:"name"`
	Topic         string `yaml:"topic"`
	Count         int    `yaml:"count"`
	RatePerSecond int    `yaml:"rate_per_second"`
	MessageSize   int    `yaml:"message_size"`
	BatchSize     int    `yaml:"batch_size,omitempty"`
}

type ConsumerGroup struct {
	Name             string `yaml:"name"`
	Topic            string `yaml:"topic"`
	Subscription     string `yaml:"subscription"`
	SubscriptionType string `yaml:"subscription_type"` // shared|exclusive|failover
	Count            int    `yaml:"count"`
	AckTimeout       string `yaml:"ack_timeout,omitempty"`
}

type MetricsConfig struct {
	Enabled        bool      `yaml:"enabled"`
	ReportInterval string    `yaml:"report_interval"`
	Percentiles    []float64 `yaml:"percentiles"`
	OutputFormat   string    `yaml:"output_format"` // terminal|json|prometheus
	ExportPath     string    `yaml:"export_path"`
	Collect        []string  `yaml:"collect"`
}
