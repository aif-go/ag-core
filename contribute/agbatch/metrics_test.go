package agbatch

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewBatchMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewBatchMetrics(&MetricsConfig{
		Namespace: "test",
		Subsystem: "batch",
		Registry:  registry,
	})

	if metrics == nil {
		t.Fatal("expected non-nil metrics")
	}
	if metrics.JobsStarted == nil {
		t.Error("JobsStarted is nil")
	}
	if metrics.StepDuration == nil {
		t.Error("StepDuration is nil")
	}

	// Verify metrics can be collected
	metrics.RecordJobStart("testJob")
	metrics.RecordJobComplete("testJob", 100*time.Millisecond)

	exec := NewStepExecution(1, "testStep")
	exec.ReadCount = 10
	exec.WriteCount = 8
	exec.SkipCount = 1
	exec.RetryCount = 1
	metrics.RecordStep("testJob", "testStep", exec, 50*time.Millisecond)

	metrics.RecordChunk(5 * time.Millisecond)

	// Gather and verify
	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}
	if len(families) == 0 {
		t.Error("expected at least one metric family")
	}
}

func TestNewBatchMetrics_Defaults(t *testing.T) {
	// Test with nil config — should use defaults
	registry := prometheus.NewRegistry()
	metrics := NewBatchMetrics(&MetricsConfig{
		Registry: registry,
	})
	metrics.RecordJobStart("defaultJob")
}

func TestNoopMetrics(t *testing.T) {
	m := NoopMetrics{}
	// Should not panic
	m.RecordJobStart("test")
	m.RecordJobComplete("test", time.Second)
	m.RecordJobFailure("test", time.Second)
	m.RecordStep("test", "step", nil, time.Second)
	m.RecordChunk(time.Millisecond)
}

func TestMetricJobListener(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewBatchMetrics(&MetricsConfig{
		Namespace: "test",
		Registry:  registry,
	})

	// Test success path
	listener := NewMetricJobListener(metrics)
	exec := NewJobExecution(1, "successJob")

	if err := listener.BeforeJob(nil, exec); err != nil {
		t.Fatal(err)
	}
	exec.Status = StatusCompleted
	if err := listener.AfterJob(nil, exec); err != nil {
		t.Fatal(err)
	}

	// Test failure path
	exec2 := NewJobExecution(2, "failJob")
	if err := listener.BeforeJob(nil, exec2); err != nil {
		t.Fatal(err)
	}
	exec2.Status = StatusFailed
	if err := listener.AfterJob(nil, exec2); err != nil {
		t.Fatal(err)
	}
}

func TestMetricStepListener(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewBatchMetrics(&MetricsConfig{
		Namespace: "test",
		Registry:  registry,
	})

	listener := NewMetricStepListener(metrics)
	exec := NewStepExecution(1, "testStep")
	exec.ReadCount = 5
	exec.WriteCount = 5

	if err := listener.BeforeStep(nil, exec); err != nil {
		t.Fatal(err)
	}
	if err := listener.AfterStep(nil, exec); err != nil {
		t.Fatal(err)
	}
}

func TestBatchMetrics_ImplementsCollector(t *testing.T) {
	// Compile-time check that BatchMetrics satisfies MetricsCollector
	var _ MetricsCollector = (*BatchMetrics)(nil)
	var _ MetricsCollector = NoopMetrics{}
}
