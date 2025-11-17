package metrics

import (
	"testing"
)

func TestRecordSeq_LossAndDuplicatesSingleKey(t *testing.T) {
	c := NewCollector()
	// Simulate messages from topic T, subscription S, producer P
	topic, sub, prod := "/default/test", "sub", "producer-1"
	// Seen: 1,2,3,5 (missing 4 => loss=1), and duplicate 3 once => duplicates=1
	c.RecordSeq(topic, sub, prod, 1)
	c.RecordSeq(topic, sub, prod, 2)
	c.RecordSeq(topic, sub, prod, 3)
	c.RecordSeq(topic, sub, prod, 3) // duplicate
	c.RecordSeq(topic, sub, prod, 5)

	snap := c.Snapshot()
	if snap.EstimatedLoss != 1 {
		t.Fatalf("EstimatedLoss got %d want 1", snap.EstimatedLoss)
	}
	if snap.Duplicates != 1 {
		t.Fatalf("Duplicates got %d want 1", snap.Duplicates)
	}
	if len(snap.IntegrityBreakdown) != 1 {
		t.Fatalf("breakdown size got %d want 1", len(snap.IntegrityBreakdown))
	}
	entry := snap.IntegrityBreakdown[0]
	if entry.Topic != topic || entry.Subscription != sub || entry.Producer != prod {
		t.Fatalf("unexpected key in breakdown: %+v", entry)
	}
	if entry.Min != 1 || entry.Max != 5 {
		t.Fatalf("range got [%d..%d] want [1..5]", entry.Min, entry.Max)
	}
	if entry.UniqueSeen != 4 { // 1,2,3,5
		t.Fatalf("unique seen got %d want 4", entry.UniqueSeen)
	}
	if entry.Loss != 1 || entry.Duplicates != 1 {
		t.Fatalf("entry loss/dup got %d/%d want 1/1", entry.Loss, entry.Duplicates)
	}
}

func TestRecordSeq_IsolatedKeysDoNotInterfere(t *testing.T) {
	c := NewCollector()
	// Two producers on same topic/subscription
	topic, sub := "/default/test", "sub"
	c.RecordSeq(topic, sub, "producer-A", 1)
	c.RecordSeq(topic, sub, "producer-A", 2)
	c.RecordSeq(topic, sub, "producer-B", 10)
	c.RecordSeq(topic, sub, "producer-B", 12) // missing 11 => loss=1 for B

	snap := c.Snapshot()
	if snap.EstimatedLoss != 1 {
		t.Fatalf("EstimatedLoss got %d want 1", snap.EstimatedLoss)
	}
	if snap.Duplicates != 0 {
		t.Fatalf("Duplicates got %d want 0", snap.Duplicates)
	}
	if len(snap.IntegrityBreakdown) != 2 {
		t.Fatalf("breakdown size got %d want 2", len(snap.IntegrityBreakdown))
	}
}
