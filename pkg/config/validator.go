package config

import (
	"fmt"
)

// Validate performs basic schema validation and returns a list of errors (if any).
func Validate(cfg *Config) []error {
	errs := []error{}

	if cfg.TestName == "" {
		errs = append(errs, fmt.Errorf("test_name is required"))
	}
	if cfg.Danube.ServiceURL == "" {
		errs = append(errs, fmt.Errorf("danube.service_url is required"))
	}
	if cfg.Execution.Duration == "" {
		errs = append(errs, fmt.Errorf("execution.duration is required"))
	}

	// Topics must have valid schema types
	allowedSchemas := map[string]bool{"json": true, "string": true, "int64": true, "number": true}
	allowedDispatch := map[string]bool{"": true, "non_reliable": true, "reliable": true}
	for i, t := range cfg.Topics {
		if t.Name == "" {
			errs = append(errs, fmt.Errorf("topics[%d].name is required", i))
		}
		if !allowedSchemas[t.SchemaType] {
			errs = append(errs, fmt.Errorf("topics[%d].schema_type must be one of json|string|int64|number", i))
		}
		if t.SchemaType == "json" && t.JSONSchema == "" {
			errs = append(errs, fmt.Errorf("topics[%d].json_schema is required when schema_type=json", i))
		}
		if t.Partitions < 0 {
			errs = append(errs, fmt.Errorf("topics[%d].partitions must be >= 0", i))
		}
		if !allowedDispatch[t.DispatchStrategy] {
			errs = append(errs, fmt.Errorf("topics[%d].dispatch_strategy must be one of reliable|non_reliable (or omitted)", i))
		}
	}

	// Producers
	for i, p := range cfg.Producers {
		if p.Topic == "" {
			errs = append(errs, fmt.Errorf("producers[%d].topic is required", i))
		}
		if p.Count <= 0 {
			errs = append(errs, fmt.Errorf("producers[%d].count must be > 0", i))
		}
		if p.RatePerSecond < 0 {
			errs = append(errs, fmt.Errorf("producers[%d].rate_per_second must be >= 0", i))
		}
		if p.MessageSize < 0 {
			errs = append(errs, fmt.Errorf("producers[%d].message_size must be >= 0", i))
		}
	}

	// Consumers
	allowedSubs := map[string]bool{"shared": true, "exclusive": true, "failover": true}
	for i, c := range cfg.Consumers {
		if c.Topic == "" {
			errs = append(errs, fmt.Errorf("consumers[%d].topic is required", i))
		}
		if c.Subscription == "" {
			errs = append(errs, fmt.Errorf("consumers[%d].subscription is required", i))
		}
		if !allowedSubs[c.SubscriptionType] {
			errs = append(errs, fmt.Errorf("consumers[%d].subscription_type must be one of shared|exclusive|failover", i))
		}
		if c.Count <= 0 {
			errs = append(errs, fmt.Errorf("consumers[%d].count must be > 0", i))
		}
	}

	// Metrics
	if cfg.Metrics.Enabled {
		if cfg.Metrics.ReportInterval == "" {
			errs = append(errs, fmt.Errorf("metrics.report_interval is required when metrics.enabled=true"))
		}
		if cfg.Metrics.OutputFormat == "" {
			errs = append(errs, fmt.Errorf("metrics.output_format is required when metrics.enabled=true"))
		}
	}

	return errs
}
