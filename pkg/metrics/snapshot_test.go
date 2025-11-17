package metrics

import (
	"encoding/json"
	"testing"
)

func TestSnapshotJSONKeys(t *testing.T) {
	b, err := json.Marshal(Snapshot{})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Required keys (no omitempty)
	required := []string{
		"elapsed_sec",
		"messages_sent",
		"messages_received",
		"errors",
		"throughput_sent",
		"throughput_recv",
		"latency_p50_ms",
		"latency_p95_ms",
		"latency_p99_ms",
		"latency_max_ms",
		"latency_samples",
		"duplicates",
		"estimated_loss",
	}
	// Optional keys (omitempty)
	optional := map[string]struct{}{
		"integrity_breakdown": {},
	}
	allowed := make(map[string]struct{}, len(required)+len(optional))
	for _, k := range required {
		allowed[k] = struct{}{}
	}
	for k := range optional {
		allowed[k] = struct{}{}
	}

	// Check all required keys are present
	for _, k := range required {
		if _, ok := got[k]; !ok {
			t.Errorf("missing json key: %q", k)
		}
	}
	// Check no unexpected keys are present
	for k := range got {
		if _, ok := allowed[k]; !ok {
			t.Errorf("unexpected json key present: %q", k)
		}
	}
}
