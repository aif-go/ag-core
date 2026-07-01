# agbatch — Batch Processing Framework

基于 [destel/rill](https://github.com/destel/rill) 的 Go 批处理框架，对齐 Spring Batch 概念（Job/Step/ItemReader/ItemProcessor/ItemWriter），贴合现有 `ag-core` 技术栈（fx DI、slog 日志、Go generics）。

## 快速开始

```go
package main

import (
    "context"
    "io"
    "log"

    "github.com/aif-go/ag-core/contribute/agbatch"
)

func main() {
    // 1. 定义 Reader（逐条读取）
    reader := agbatch.ReaderFunc[int](func(_ context.Context) (int, error) {
        // 从数据源读取...
        return 0, io.EOF
    })

    // 2. 定义 Processor（逐条处理）
    processor := agbatch.ProcessorFunc[int, string](func(_ context.Context, item int) (string, error) {
        // 转换/验证...
        return fmt.Sprintf("id-%d", item), nil
    })

    // 3. 定义 Writer（批量写入）
    writer := agbatch.WriterFunc[string](func(_ context.Context, items []string) error {
        // 批量写入数据库...
        return nil
    })

    // 4. 构建 Step 和 Job
    job := agbatch.NewJobBuilder("myJob").
        Step(agbatch.NewChunkStepBuilder[int, string]("processStep").
            Reader(reader).
            Processor(processor).
            Writer(writer).
            ChunkSize(100).
            RetryPolicy(agbatch.MaxAttempts(3, time.Second)).
            SkipPolicy(agbatch.SkipLimit(10)).
            ProcessorPoolSize(4).
            Build(),
        ).
        Listener(agbatch.JobListenerFunc(
            func(_ context.Context, e *agbatch.JobExecution) error {
                log.Printf("Job %s starting", e.JobName)
                return nil
            },
            func(_ context.Context, e *agbatch.JobExecution) error {
                log.Printf("Job %s finished: %s", e.JobName, e.Status)
                return nil
            },
        )).
        Build()

    // 5. 启动
    launcher := agbatch.NewJobLauncher(nil)
    exec, err := launcher.Run(context.Background(), job)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Read: %d, Written: %d, Skipped: %d, Retried: %d",
        exec.StepExecs[0].ReadCount, exec.StepExecs[0].WriteCount,
        exec.StepExecs[0].SkipCount, exec.StepExecs[0].RetryCount)
}
```

## 核心概念

### Job（作业）
完整批处理过程，由多个 Step 按顺序组成。失败即停止，不执行后续 Step。

### Step（步骤）
- **ChunkStep** — Chunk 模式（read-process-write），对标 Spring Batch chunk-oriented processing
- **TaskletStep** — 单一任务（存过调用、文件清理等无需分块的操作）

### ItemReader / ItemProcessor / ItemWriter
三个核心接口，支持函数适配器快速实现：

| 接口 | 方法 | 说明 |
|------|------|------|
| `ItemReader[T]` | `Read(ctx) (T, error)` | 返回 `io.EOF` 表示结束 |
| `ItemProcessor[T,R]` | `Process(ctx, T) (R, error)` | 逐条转换 |
| `ItemWriter[T]` | `Write(ctx, []T) error` | 批量写入，切片长度 ≤ chunkSize |

### Chunk 处理
基于 rill pipeline 实现：
1. `rill.Generate` 包装 ItemReader 生成 item 流
2. `rill.Batch(chunkSize, timeout)` 将 item 分块
3. 块内并发处理（ProcessorPoolSize > 1 时使用 `rill.Map`）
4. `rill.ForEach` 执行写入

### Retry / Skip
| 策略 | 函数 | 说明 |
|------|------|------|
| 固定重试 | `MaxAttempts(n, delay)` | 最多 n 次，固定间隔 |
| 指数退避 | `ExponentialBackoff(n, init, max)` | 延迟 × 2^attempt |
| 条件重试 | `RetryableError(delegate, predicate)` | 仅特定错误重试 |
| 限制跳过 | `SkipLimit(n)` | 最多跳过 n 条 |
| 条件跳过 | `SkipOnError(predicate, limit)` | 仅跳过特定错误 |
| 永不重试/跳过 | `NoRetry()` / `NeverSkip()` | 任何错误即失败 |

### Listener（监听器）
- **JobListener** — BeforeJob / AfterJob
- **StepListener** — BeforeStep / AfterStep
- **ChunkListener** — BeforeChunk / AfterChunk / OnChunkError

### Repository
`JobRepository` 接口持久化执行元数据。默认提供线程安全的 `InMemoryRepository`。生产环境可替换为 DB 实现。

### fx 集成
```go
app := fx.New(
    agbatch.FxAgBatchModule,  // 提供 JobLauncher + InMemoryRepository
    fx.Provide(NewMyJob),
)
```

## Spring Batch 概念对照

| Spring Batch | agbatch |
|-------------|---------|
| Job | `Job` (NewJobBuilder) |
| Step | `ChunkStep` / `TaskletStep` |
| ItemReader | `ItemReader[T]` / `ReaderFunc[T]` |
| ItemProcessor | `ItemProcessor[T,R]` / `ProcessorFunc[T,R]` |
| ItemWriter | `ItemWriter[T]` / `WriterFunc[T]` |
| Chunk | `ChunkSize(n)`, `ProcessorPoolSize(n)` |
| RetryPolicy | `RetryPolicy` 接口 + `MaxAttempts`, `ExponentialBackoff` |
| SkipPolicy | `SkipPolicy` 接口 + `SkipLimit`, `SkipOnError` |
| JobRepository | `JobRepository` 接口 + `InMemoryRepository` |
| JobLauncher | `JobLauncher.Run()` |
| JobExecutionListener | `JobListener` |
| StepExecutionListener | `StepListener` |
| ChunkListener | `ChunkListener` |
| Tasklet | `Tasklet` / `TaskletFunc` |
| ExecutionContext | `ExecutionContext` |
| PartitionedStep | `PartitionedStep` + `Partitioner` 接口 |
| FlowStep / Decision | `FlowStep` + `FlowDecision` 接口 |
| JdbcCursorItemReader | `SqlCursorItemReader[T]` |
| JdbcPagingItemReader | `SqlPagingItemReader[T]` |
| JdbcBatchItemWriter | `SqlBatchItemWriter[T]` |
| HibernateCursorItemReader | `GormCursorItemReader[T]` |
| HibernatePagingItemReader | `GormPagingItemReader[T]` |
| HibernateItemWriter | `GormBatchItemWriter[T]` / `GormUpsertItemWriter[T]` |
| Metrics | `BatchMetrics` (Prometheus) + `MetricsCollector` 接口 |

## 并发模型

基于 rill 的并发模型：
- **步骤间**：顺序执行（Step 1 → Step 2 → Step 3）
- **块内处理**：通过 `ProcessorPoolSize(n)` 控制并发度
- **背压**：Go channel 原生反压，下游慢则上游自动减速
- **上下文取消**：`ctx.Done()` 检查，支持优雅停止

## 包结构

```
agbatch/                  # 核心框架（Job/Step/Retry/Listener/Repository）
├── reader/               # 内置 Reader 实现
│   ├── sql.go            # SqlCursorItemReader, SqlPagingItemReader
│   ├── gorm.go           # GormCursorItemReader, GormPagingItemReader
│   └── conditionwhere.go # FieldMask → GORM WHERE 条件过滤
└── writer/               # 内置 Writer 实现
    ├── sql.go            # SqlBatchItemWriter
    └── gorm.go           # GormBatchItemWriter, GormUpsertItemWriter
```

## 内置 SQL Reader / Writer

对标 Spring Batch 的 `JdbcCursorItemReader` / `JdbcPagingItemReader` / `JdbcBatchItemWriter`，直接基于 SQL 读写数据库，无需手写 Reader/Writer。

```go
import (
    agbatch "github.com/aif-go/ag-core/contribute/agbatch"
    agreader "github.com/aif-go/ag-core/contribute/agbatch/reader"
    agwriter "github.com/aif-go/ag-core/contribute/agbatch/writer"
)
```

### SqlCursorItemReader（游标读取）

打开一个数据库游标，逐行读取，适用于中小数据集。

```go
reader := agbatch.NewSqlCursorItemReader(db,
    "SELECT id, name, email FROM users WHERE status = ?", 1,
    func(rows *sql.Rows) (*User, error) {
        var u User
        return &u, rows.Scan(&u.ID, &u.Name, &u.Email)
    },
)
defer reader.Close()
```

### SqlPagingItemReader（分页读取）

分页查询，每页释放连接，适用于大数据集。查询中 `{limit}` 和 `{offset}` 自动替换。

```go
reader := agbatch.NewSqlPagingItemReader(db,
    "SELECT id, name FROM users ORDER BY id LIMIT {limit} OFFSET {offset}",
    500, // page size
    func(rows *sql.Rows) (*User, error) {
        var u User
        return &u, rows.Scan(&u.ID, &u.Name)
    },
)
reader.Reset() // 重置到第一页，支持重启
```

### SqlBatchItemWriter（批量写入）

在一个事务中批量执行 SQL，出错自动回滚。

```go
writer := agbatch.NewSqlBatchItemWriter(db,
    "INSERT INTO users (id, name) VALUES (?, ?)",
    func(u *User) []any { return []any{u.ID, u.Name} },
)
```

### 与 GORM / agdb 结合

现有 GORM DAO 可直接获取 `*sql.DB` 使用：

```go
sqlDB, _ := gormDB.DB()
reader := agbatch.NewSqlCursorItemReader(sqlDB, "SELECT ...", rowMapper)
writer := agbatch.NewSqlBatchItemWriter(sqlDB, "INSERT ...", preparer)
```

更推荐直接用 GORM 版本，连 SQL 都不用写：

### GormCursorItemReader（GORM 游标读取）

```go
reader := agbatch.NewGormCursorItemReader[User](db, func(db *gorm.DB) *gorm.DB {
    return db.Model(&User{}).Where("status = ?", "active").Order("id")
})
defer reader.Close()
```

### GormPagingItemReader（GORM 分页读取）

```go
reader := agbatch.NewGormPagingItemReader[User](db, 500,
    func(db *gorm.DB) *gorm.DB {
        return db.Model(&User{}).Order("id")
    },
)
reader.Reset()
```

### GormBatchItemWriter（GORM 批量写入）

```go
writer := agbatch.NewGormBatchItemWriter[User](db, 100) // CreateInBatches(100)
```

### GormUpsertItemWriter（GORM 批量 upsert）

```go
writer := agbatch.NewGormUpsertItemWriter[User](db,
    clause.OnConflict{UpdateAll: true},
)
```

## 高级特性

### DB-backed JobRepository

`SqlJobRepository` 将执行元数据持久化到 SQL 数据库（MySQL/Postgres/SQLite），支持重启后查询历史执行记录。

```go
db, _ := sql.Open("mysql", dsn)
repo := agbatch.NewSqlJobRepository(db, "mysql")
job := agbatch.NewJobBuilder("persistentJob").Repository(repo).Step(...).Build()
```

### PartitionedStep（分区并行）

将 Reader 的数据按分区策略分配到 N 个分区，每个分区独立并发处理。

```go
step := agbatch.NewPartitionedStep(agbatch.PartitionedStepConfig{
    Name:          "partitionedStep",
    Reader:        reader,
    Partitioner:   agbatch.RangePartitioner[int]{},
    Processor:     processor,
    Writer:        writer,
    NumPartitions: 8,
    ChunkSize:     100,
})
```

内置三种分区策略：`RoundRobinPartitioner`、`RangePartitioner`、`HashPartitioner`。

### FlowStep（条件分支）

支持 Step 间的条件跳转，根据执行状态或退出码决定下一步。

```go
flow := agbatch.NewFlowStepBuilder("decisionFlow").
    First("validate").
    Decider("validate", validateStep, agbatch.OnStatus(map[agbatch.BatchStatus]string{
        agbatch.StatusCompleted: "process",
        agbatch.StatusFailed:    "reportFailure",
    })).
    Decider("process", processStep, agbatch.OnExitCode(map[string]string{
        "CONTINUE": "enrich",
        "COMPLETED": "",
    })).
    Step("enrich", enrichStep).
    Step("reportFailure", reportStep).
    Build()
```

### Prometheus Metrics

自动收集 Job/Step/Chunk 级别指标：执行时长、读写跳过计数、活跃任务数。

```go
metrics := agbatch.NewBatchMetrics(&agbatch.MetricsConfig{Namespace: "myapp"})
job := agbatch.NewJobBuilder("monitoredJob").
    Listener(agbatch.NewMetricJobListener(metrics)).
    Step(agbatch.NewChunkStepBuilder[...]("step").
        Listener(agbatch.NewMetricStepListener(metrics)).
        Build(),
    ).Build()
```

暴露的指标：`myapp_jobs_started_total`, `myapp_jobs_completed_total`, `myapp_step_duration_seconds`, `myapp_items_read_total` 等。
