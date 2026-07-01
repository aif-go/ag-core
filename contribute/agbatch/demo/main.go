// Package main demonstrates all agbatch features in a realistic ETL pipeline:
// orders are read from a source DB, enriched, validated, and written to a destination DB.
//
// Features covered:
//   - ChunkStep (read-process-write) with retry & skip
//   - TaskletStep (setup / cleanup / report)
//   - PartitionedStep (parallel processing)
//   - FlowStep (conditional branching)
//   - SqlCursorItemReader / SqlBatchItemWriter (raw SQL)
//   - GormCursorItemReader / GormBatchItemWriter (GORM)
//   - GormUpsertItemWriter (ON CONFLICT upsert)
//   - RetryPolicy (exponential backoff with max attempts)
//   - SkipPolicy (skip on specific errors)
//   - JobListener / StepListener (lifecycle callbacks)
//   - Prometheus Metrics (BatchMetrics)
//   - conditonwhere FieldMask (dynamic WHERE filtering)
//   - ExecutionContext (data sharing between steps)
//   - InMemoryRepository & SqlJobRepository
//
// Run:
//
//	go run ./contribute/agbatch/demo/
package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	agbatch "github.com/aif-go/ag-core/contribute/agbatch"
	agreader "github.com/aif-go/ag-core/contribute/agbatch/reader"
	agstep "github.com/aif-go/ag-core/contribute/agbatch/step"
	agwriter "github.com/aif-go/ag-core/contribute/agbatch/writer"

	cw "github.com/aif-go/ag-core/contribute/agdb/conditonwhere"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"net/http"

	_ "modernc.org/sqlite"
)

// ========== Domain Types ==========

// Order represents a source order that needs processing.
type Order struct {
	ID       int    `gorm:"primaryKey"`
	Customer string `gorm:"size:100"`
	Amount   float64
	Status   string `gorm:"size:20"` // "new", "processing", "done", "invalid"
	Region   string `gorm:"size:10"` // "EAST", "WEST", "NORTH", "SOUTH"
}

// EnrichedOrder is the processed/validated order written to the destination.
type EnrichedOrder struct {
	ID          int    `gorm:"primaryKey"`
	Customer    string `gorm:"size:100"`
	Amount      float64
	Tax         float64
	Total       float64
	Region      string `gorm:"size:10"`
	ProcessedAt string `gorm:"size:30"`
}

// ========== In-Memory Database Setup ==========

func setupSourceDB() *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}
	db.Exec(`CREATE TABLE orders (
		id INTEGER PRIMARY KEY,
		customer TEXT, amount REAL, status TEXT, region TEXT
	)`)

	// Insert 100 source orders with some variety
	regions := []string{"EAST", "WEST", "NORTH", "SOUTH"}
	statuses := []string{"new", "new", "new", "new", "new", "invalid"} // 1/6 invalid
	for i := range 100 {
		id := i + 1
		customer := fmt.Sprintf("customer-%d", id)
		amount := float64(50+rand.Intn(450)) + float64(rand.Intn(100))/100.0
		status := statuses[rand.Intn(len(statuses))]
		region := regions[rand.Intn(len(regions))]
		db.Exec("INSERT INTO orders VALUES (?, ?, ?, ?, ?)", id, customer, amount, status, region)
	}
	return db
}

func setupDestDB() *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}
	db.Exec(`CREATE TABLE enriched_orders (
		id INTEGER PRIMARY KEY,
		customer TEXT, amount REAL, tax REAL, total REAL, region TEXT, processed_at TEXT
	)`)
	return db
}

func setupGormDestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("agbatch-demo-gorm.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.Exec("PRAGMA journal_mode=WAL") // enable concurrent reads during writes
	db.Exec("DROP TABLE IF EXISTS enriched_orders")
	db.AutoMigrate(&EnrichedOrder{})
	return db
}

// ========== Helpers ==========

func DBGetter(db *gorm.DB) func(context.Context) *gorm.DB {
	return func(ctx context.Context) *gorm.DB { return db.WithContext(ctx) }
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func ptr[T any](v T) *T { return &v }

// ========== Main ==========

func main() {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  AGBATCH COMPREHENSIVE DEMO")
	fmt.Println(strings.Repeat("=", 70))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// ---- Setup databases ----
	srcDB := setupSourceDB()
	destDB := setupDestDB()
	gormDestDB := setupGormDestDB()

	// ---- Metrics ----
	metricsRegistry := prometheus.NewRegistry()
	metrics := agbatch.NewBatchMetrics(&agbatch.MetricsConfig{
		Namespace: "agbatch_demo",
		Registry:  metricsRegistry,
	})

	// Start metrics HTTP server
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{}))
		fmt.Println("  Metrics: http://localhost:2112/metrics")
		http.ListenAndServe(":2112", mux)
	}()

	var totalProcessed atomic.Int64

	// ===================================================================
	// BUILD THE JOB
	// ===================================================================
	job := agbatch.NewJobBuilder("order-etl-pipeline").

		// ── Step 1: Tasklet — Setup ──────────────────────────────────
		Step(agbatch.NewTaskletStep("setup",
			agbatch.TaskletFunc(func(ctx context.Context, exec *agbatch.StepExecution) error {
				fmt.Println("\n  [step:setup] Initializing ETL pipeline...")
				exec.Context.Set("start_time", time.Now().Format(time.RFC3339))
				exec.Context.Set("batch_id", fmt.Sprintf("batch-%d", time.Now().Unix()))
				fmt.Printf("  [step:setup] Batch ID: %s\n", exec.Context.Get("batch_id"))
				return nil
			}),
			nil, // no step listener for this step
		)).

		// ── Step 2: ChunkStep — SQL Reader/Writer with Retry & Skip ──
		Step(agstep.NewChunkStepBuilder[*Order, *EnrichedOrder]("process-orders").
			Reader(agreader.NewSqlCursorItemReader(srcDB,
				"SELECT id, customer, amount, status, region FROM orders WHERE status = ?",
				func(rows *sql.Rows) (*Order, error) {
					var o Order
					return &o, rows.Scan(&o.ID, &o.Customer, &o.Amount, &o.Status, &o.Region)
				},
				"new", // only process "new" orders
			)).
			Processor(agbatch.ProcessorFunc[*Order, *EnrichedOrder](
				func(ctx context.Context, order *Order) (*EnrichedOrder, error) {
					// Simulate transient failures for retry demo (1 in 10 chance)
					if rand.Intn(10) == 0 {
						return nil, fmt.Errorf("transient processor error for order %d", order.ID)
					}
					// Simulate invalid data for skip demo (orders with amount < 0 don't exist,
					// but we check a condition)
					if order.Amount <= 0 {
						return nil, fmt.Errorf("invalid amount: %v", order.Amount)
					}

					tax := order.Amount * 0.08
					return &EnrichedOrder{
						ID:          order.ID,
						Customer:    strings.ToUpper(order.Customer),
						Amount:      order.Amount,
						Tax:         tax,
						Total:       order.Amount + tax,
						Region:      order.Region,
						ProcessedAt: time.Now().Format(time.RFC3339),
					}, nil
				},
			)).
			Writer(agwriter.NewSqlBatchItemWriter(destDB,
				"INSERT INTO enriched_orders (id, customer, amount, tax, total, region, processed_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
				func(e *EnrichedOrder) []any {
					return []any{e.ID, e.Customer, e.Amount, e.Tax, e.Total, e.Region, e.ProcessedAt}
				},
			)).
			ChunkSize(10).
			ProcessorPoolSize(4).
			// Retry: up to 3 attempts with exponential backoff
			RetryPolicy(agbatch.ExponentialBackoff(3, 50*time.Millisecond, 500*time.Millisecond)).
			// Skip: up to 5 items with "invalid" in the error
			SkipPolicy(agbatch.SkipOnError(
				func(err error) bool { return strings.Contains(err.Error(), "invalid") },
				5,
			)).
			Listener(agbatch.NewMetricStepListener(metrics)).
			Build(),
		).

		// ── Step 3: PartitionedStep — Parallel partition processing with GORM ──
		Step(agstep.NewPartitionedStep(agstep.PartitionedStepConfig{
			Name: "partition-enrich",
			Reader: &anyReader{delegate: agreader.NewSqlCursorItemReader(destDB,
				"SELECT id, customer, amount, tax, total, region, processed_at FROM enriched_orders",
				func(rows *sql.Rows) (EnrichedOrder, error) {
					var e EnrichedOrder
					return e, rows.Scan(&e.ID, &e.Customer, &e.Amount, &e.Tax, &e.Total, &e.Region, &e.ProcessedAt)
				},
			)},
			Partitioner: &anyPartitioner{delegate: RangePartitioner{}},
			Processor: &anyProcessor{delegate: agbatch.ProcessorFunc[EnrichedOrder, EnrichedOrder](
				func(ctx context.Context, e EnrichedOrder) (EnrichedOrder, error) {
					// Add region-based bonus multiplier
					switch e.Region {
					case "EAST":
						e.Total *= 1.05
					case "WEST":
						e.Total *= 1.03
					}
					return e, nil
				},
			)},
			Writer: &anyWriter{delegate: agwriter.NewGormUpsertItemWriter[EnrichedOrder](DBGetter(gormDestDB),
				clause.OnConflict{UpdateAll: true},
			)},
			NumPartitions: 4,
			ChunkSize:     10,
			Listener:      agbatch.NewMetricStepListener(metrics),
		})).

		// ── Step 4: FlowStep — Conditional branching ──────────────────
		Step(buildFlowStep(&totalProcessed, gormDestDB, metrics)).

		// ── Step 5: Tasklet — Report ──────────────────────────────────
		Step(agbatch.NewTaskletStep("report",
			agbatch.TaskletFunc(func(ctx context.Context, exec *agbatch.StepExecution) error {
				fmt.Println("\n  [step:report] ========== Generating Report ==========")
				// Access data from JobExecution context (shared across steps)
				fmt.Printf("  [step:report] Total processed : %d\n", totalProcessed.Load())

				// Check how many rows are in each destination
				var srcCount, destCount, gormCount int
				srcDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders WHERE status='new'").Scan(&srcCount)
				destDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM enriched_orders").Scan(&destCount)
				gormDestDB.WithContext(ctx).Model(&EnrichedOrder{}).Count(ptr(int64(0)))
				gormDestDB.WithContext(ctx).Raw("SELECT COUNT(*) FROM enriched_orders").Scan(&gormCount)

				fmt.Printf("  [step:report] Source (new orders) : %d\n", srcCount)
				fmt.Printf("  [step:report] Dest (SQL writer)    : %d\n", destCount)
				fmt.Printf("  [step:report] Dest (GORM upsert)   : %d\n", gormCount)
				return nil
			}),
			nil,
		)).

		// ── Job Listener ─────────────────────────────────────────────
		Listener(agbatch.NewMetricJobListener(metrics)).
		Build()

	// ===================================================================
	// LAUNCH
	// ===================================================================
	fmt.Println("\n  Launching job...")
	launcher := agbatch.NewJobLauncher(agbatch.NewInMemoryRepository())
	exec, err := launcher.Run(ctx, job)
	if err != nil {
		fmt.Printf("\n  ❌ Job FAILED: %v\n", err)
		os.Exit(1)
	}

	// ===================================================================
	// RESULTS
	// ===================================================================
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  JOB COMPLETED")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("  Job Name   : %s\n", exec.JobName)
	fmt.Printf("  Status     : %s\n", exec.Status)
	fmt.Printf("  Duration   : %s\n", exec.EndTime.Sub(exec.StartTime).Round(time.Millisecond))
	fmt.Println()

	for i, se := range exec.StepExecs {
		fmt.Printf("  Step %d: %-20s  status=%-10s  read=%-5d write=%-5d skip=%-5d retry=%-5d\n",
			i+1, se.StepName, se.Status, se.ReadCount, se.WriteCount, se.SkipCount, se.RetryCount)
	}

	fmt.Println()
	fmt.Println("  Metrics available at: http://localhost:2112/metrics")
	fmt.Println("  (Press Ctrl+C to exit)")

	// Wait for signal
	<-ctx.Done()
	fmt.Println("\n  Shutting down...")
}

// ========== FlowStep Builder ==========

func buildFlowStep(totalProcessed *atomic.Int64, gormDB *gorm.DB, metrics *agbatch.BatchMetrics) *agstep.FlowStep {

	// Simulate an enrichment step that may fail for some records
	validateStep := agbatch.NewTaskletStep("validate",
		agbatch.TaskletFunc(func(ctx context.Context, exec *agbatch.StepExecution) error {
			var count int64
			gormDB.WithContext(ctx).Model(&EnrichedOrder{}).Count(&count)
			fmt.Printf("  [flow:validate] Found %d enriched orders\n", count)

			if count == 0 {
				exec.Status = agbatch.StatusFailed
				exec.ExitStatus = &agbatch.ExitStatus{Code: "NO_DATA"}
				return fmt.Errorf("no orders to validate")
			}

			totalProcessed.Store(count)
			exec.Status = agbatch.StatusCompleted
			exec.ExitStatus = &agbatch.ExitStatus{Code: "COMPLETED"}
			exec.Context.Set("validated_count", count)
			return nil
		}),
		nil,
	)

	errorHandler := agbatch.NewTaskletStep("error-handler",
		agbatch.TaskletFunc(func(ctx context.Context, exec *agbatch.StepExecution) error {
			fmt.Println("  [flow:error-handler] Handling error gracefully...")
			fmt.Printf("  [flow:error-handler] Previous step exit code: %s\n",
				jobExecStepExitCode(exec))
			return nil
		}),
		nil,
	)

	// Apply conditonwhere FieldMask to filter records for final enrichment
	enrichStep := agstep.NewChunkStepBuilder[*EnrichedOrder, *EnrichedOrder]("final-enrich").
		Reader(agreader.NewGormCursorItemReader[*EnrichedOrder](DBGetter(gormDB),
			agreader.GormConditionWhereQuery(
				func(db *gorm.DB) *gorm.DB { return db.Model(&EnrichedOrder{}).Order("id") },
				"region = @Region AND total > @MinTotal",
				agreader.FieldMaskFromMap("Region", "MinTotal"),
				map[string]any{"Region": "EAST", "MinTotal": 100.0},
			),
		)).
		Processor(agbatch.ProcessorFunc[*EnrichedOrder, *EnrichedOrder](
			func(ctx context.Context, e *EnrichedOrder) (*EnrichedOrder, error) {
				e.Customer = "PRIORITY-" + e.Customer
				return e, nil
			},
		)).
		Writer(agwriter.NewGormUpsertItemWriter[*EnrichedOrder](DBGetter(gormDB),
			clause.OnConflict{UpdateAll: true},
		)).
		ChunkSize(5).
		Listener(agbatch.NewMetricStepListener(metrics)).
		Build()

	return agstep.NewFlowStepBuilder("quality-gate").
		First("validate").
		Decider("validate", validateStep, buildQualityDecision()).
		Step("final-enrich", enrichStep).    // terminal after enrich
		Step("error-handler", errorHandler). // terminal
		Build()
}

func buildQualityDecision() agbatch.FlowDecision {
	return agbatch.OnStatus(map[agbatch.BatchStatus]string{
		agbatch.StatusCompleted: "final-enrich",
		agbatch.StatusFailed:    "error-handler",
	})
}

func jobExecStepExitCode(exec *agbatch.StepExecution) string {
	if exec.ExitStatus != nil {
		return exec.ExitStatus.Code
	}
	return "UNKNOWN"
}

// ========== Internal Type Adapters (for PartitionedStep) ==========

type anyReader struct {
	delegate agbatch.ItemReader[EnrichedOrder]
}

func (r *anyReader) Read(ctx context.Context) (any, error) { return r.delegate.Read(ctx) }

type anyProcessor struct {
	delegate agbatch.ItemProcessor[EnrichedOrder, EnrichedOrder]
}

func (p *anyProcessor) Process(ctx context.Context, item any) (any, error) {
	return p.delegate.Process(ctx, item.(EnrichedOrder))
}

type anyWriter struct {
	delegate agbatch.ItemWriter[EnrichedOrder]
}

func (w *anyWriter) Write(ctx context.Context, items []any) error {
	typed := make([]EnrichedOrder, len(items))
	for i, item := range items {
		if item != nil {
			typed[i] = item.(EnrichedOrder)
		}
	}
	return w.delegate.Write(ctx, typed)
}

type RangePartitioner struct{}

func (RangePartitioner) Partition(items []EnrichedOrder, n int) [][]EnrichedOrder {
	if n <= 0 {
		n = 1
	}
	if n > len(items) {
		n = len(items)
	}
	parts := make([][]EnrichedOrder, n)
	chunkSize := (len(items) + n - 1) / n
	for i := range n {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(items) {
			end = len(items)
		}
		if start < end {
			parts[i] = items[start:end]
		}
	}
	return parts
}

type anyPartitioner struct {
	delegate interface {
		Partition(items []EnrichedOrder, n int) [][]EnrichedOrder
	}
}

func (p *anyPartitioner) Partition(items []any, n int) [][]any {
	typed := make([]EnrichedOrder, len(items))
	for i, item := range items {
		typed[i] = item.(EnrichedOrder)
	}
	parts := p.delegate.Partition(typed, n)
	result := make([][]any, len(parts))
	for i, part := range parts {
		result[i] = make([]any, len(part))
		for j, v := range part {
			result[i][j] = v
		}
	}
	return result
}

// ========== Ensure Imports Used ==========
var _ = slog.Default
var _ = io.EOF
var _ = prometheus.DefaultRegisterer
var _ = sqlite.Open
var _ = gorm.Expr
var _ = clause.Expression(nil)
var _ = cw.FieldMask{}
