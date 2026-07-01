package agbatch

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// BatchMetrics collects operational metrics for batch processing.
// It wraps Prometheus counters, histograms, and gauges for job/step/chunk monitoring.
type BatchMetrics struct {
	JobsStarted   prometheus.Counter
	JobsCompleted prometheus.Counter
	JobsFailed    prometheus.Counter
	JobDuration   prometheus.Histogram

	// StepDuration is labeled by job_name and step_name.
	StepDuration *prometheus.HistogramVec

	ChunkDuration prometheus.Histogram

	ItemsRead    prometheus.Counter
	ItemsWritten prometheus.Counter
	ItemsSkipped prometheus.Counter
	ItemsRetried prometheus.Counter

	ActiveJobs  prometheus.Gauge
	ActiveSteps prometheus.Gauge
}

// MetricsConfig configures metric collection.
type MetricsConfig struct {
	Namespace string
	Subsystem string
	// Registry is the prometheus registerer. nil = prometheus.DefaultRegisterer.
	Registry prometheus.Registerer
}

// NewBatchMetrics creates a new metrics collector.
func NewBatchMetrics(cfg *MetricsConfig) *BatchMetrics {
	if cfg == nil {
		cfg = &MetricsConfig{}
	}
	if cfg.Namespace == "" {
		cfg.Namespace = "agbatch"
	}

	reg := cfg.Registry
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	factory := promauto.With(reg)

	stepLabels := []string{"job_name", "step_name"}

	return &BatchMetrics{
		JobsStarted: factory.NewCounter(prometheus.CounterOpts{
			Namespace: cfg.Namespace, Subsystem: cfg.Subsystem,
			Name: "jobs_started_total",
			Help: "Total number of batch jobs started.",
		}),
		JobsCompleted: factory.NewCounter(prometheus.CounterOpts{
			Namespace: cfg.Namespace, Subsystem: cfg.Subsystem,
			Name: "jobs_completed_total",
			Help: "Total number of batch jobs completed successfully.",
		}),
		JobsFailed: factory.NewCounter(prometheus.CounterOpts{
			Namespace: cfg.Namespace, Subsystem: cfg.Subsystem,
			Name: "jobs_failed_total",
			Help: "Total number of batch jobs that failed.",
		}),
		JobDuration: factory.NewHistogram(prometheus.HistogramOpts{
			Namespace: cfg.Namespace, Subsystem: cfg.Subsystem,
			Name:    "job_duration_seconds",
			Help:    "Duration of batch job execution in seconds.",
			Buckets: []float64{0.1, 0.5, 1, 5, 10, 30, 60, 300, 600, 1800, 3600},
		}),
		StepDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: cfg.Namespace, Subsystem: cfg.Subsystem,
			Name:    "step_duration_seconds",
			Help:    "Duration of step execution in seconds.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10, 30, 60, 300, 600},
		}, stepLabels),
		ChunkDuration: factory.NewHistogram(prometheus.HistogramOpts{
			Namespace: cfg.Namespace, Subsystem: cfg.Subsystem,
			Name:    "chunk_duration_seconds",
			Help:    "Duration of chunk processing in seconds.",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10},
		}),
		ItemsRead: factory.NewCounter(prometheus.CounterOpts{
			Namespace: cfg.Namespace, Subsystem: cfg.Subsystem,
			Name: "items_read_total",
			Help: "Total number of items read.",
		}),
		ItemsWritten: factory.NewCounter(prometheus.CounterOpts{
			Namespace: cfg.Namespace, Subsystem: cfg.Subsystem,
			Name: "items_written_total",
			Help: "Total number of items written.",
		}),
		ItemsSkipped: factory.NewCounter(prometheus.CounterOpts{
			Namespace: cfg.Namespace, Subsystem: cfg.Subsystem,
			Name: "items_skipped_total",
			Help: "Total number of items skipped.",
		}),
		ItemsRetried: factory.NewCounter(prometheus.CounterOpts{
			Namespace: cfg.Namespace, Subsystem: cfg.Subsystem,
			Name: "items_retried_total",
			Help: "Total number of items retried.",
		}),
		ActiveJobs: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: cfg.Namespace, Subsystem: cfg.Subsystem,
			Name: "active_jobs",
			Help: "Number of currently active batch jobs.",
		}),
		ActiveSteps: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: cfg.Namespace, Subsystem: cfg.Subsystem,
			Name: "active_steps",
			Help: "Number of currently active batch steps.",
		}),
	}
}

// MetricsCollector records batch execution metrics. Implement with Prometheus, statsd, etc.
type MetricsCollector interface {
	RecordJobStart(jobName string)
	RecordJobComplete(jobName string, duration time.Duration)
	RecordJobFailure(jobName string, duration time.Duration)
	RecordStep(jobName, stepName string, exec *StepExecution, duration time.Duration)
	RecordChunk(duration time.Duration)
}

var _ MetricsCollector = (*BatchMetrics)(nil)

func (m *BatchMetrics) RecordJobStart(_ string) {
	m.JobsStarted.Inc()
	m.ActiveJobs.Inc()
}

func (m *BatchMetrics) RecordJobComplete(_ string, duration time.Duration) {
	m.JobsCompleted.Inc()
	m.JobDuration.Observe(duration.Seconds())
	m.ActiveJobs.Dec()
}

func (m *BatchMetrics) RecordJobFailure(_ string, duration time.Duration) {
	m.JobsFailed.Inc()
	m.JobDuration.Observe(duration.Seconds())
	m.ActiveJobs.Dec()
}

func (m *BatchMetrics) RecordStep(jobName, stepName string, exec *StepExecution, duration time.Duration) {
	m.StepDuration.WithLabelValues(jobName, stepName).Observe(duration.Seconds())
	if exec != nil {
		m.ItemsRead.Add(float64(exec.ReadCount))
		m.ItemsWritten.Add(float64(exec.WriteCount))
		m.ItemsSkipped.Add(float64(exec.SkipCount))
		m.ItemsRetried.Add(float64(exec.RetryCount))
	}
}

func (m *BatchMetrics) RecordChunk(duration time.Duration) {
	m.ChunkDuration.Observe(duration.Seconds())
}

// NoopMetrics is a no-op metrics collector.
type NoopMetrics struct{}

func (NoopMetrics) RecordJobStart(_ string)                                   {}
func (NoopMetrics) RecordJobComplete(_ string, _ time.Duration)               {}
func (NoopMetrics) RecordJobFailure(_ string, _ time.Duration)                {}
func (NoopMetrics) RecordStep(_, _ string, _ *StepExecution, _ time.Duration) {}
func (NoopMetrics) RecordChunk(_ time.Duration)                               {}

// --- Convenience listeners ---

// NewMetricJobListener creates a JobListener that records metrics.
func NewMetricJobListener(metrics MetricsCollector) JobListener {
	return &metricJobListener{metrics: metrics}
}

type metricJobListener struct {
	noopJobListener
	metrics MetricsCollector
	start   time.Time
}

func (l *metricJobListener) BeforeJob(_ context.Context, exec *JobExecution) error {
	l.start = time.Now()
	l.metrics.RecordJobStart(exec.JobName)
	return nil
}

func (l *metricJobListener) AfterJob(_ context.Context, exec *JobExecution) error {
	d := time.Since(l.start)
	if exec.Status == StatusCompleted {
		l.metrics.RecordJobComplete(exec.JobName, d)
	} else {
		l.metrics.RecordJobFailure(exec.JobName, d)
	}
	return nil
}

// NewMetricStepListener creates a StepListener that records metrics.
func NewMetricStepListener(metrics MetricsCollector) StepListener {
	return &metricStepListener{metrics: metrics}
}

type metricStepListener struct {
	noopStepListener
	metrics MetricsCollector
	start   time.Time
}

func (l *metricStepListener) BeforeStep(_ context.Context, _ *StepExecution) error {
	l.start = time.Now()
	return nil
}

func (l *metricStepListener) AfterStep(_ context.Context, exec *StepExecution) error {
	d := time.Since(l.start)
	l.metrics.RecordStep("", exec.StepName, exec, d)
	return nil
}

// Ensure context import is used.
var _ = context.Background
