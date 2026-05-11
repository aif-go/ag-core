package async

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// WorkerGroup 独立管理 worker 池和队列
type WorkerGroup struct {
	config *AsyncGroupConfig

	// 日志队列
	logQueue chan *logTask

	// worker 池
	workers     []*worker
	workerCount atomic.Int64

	// 控制通道
	shutdownChan chan struct{}
	doneChan     chan struct{}

	// 统计信息
	stats *WorkerStats

	// 引用计数
	refCount atomic.Int64

	mu sync.RWMutex
}

type WorkerStats struct {
	Queued    atomic.Int64
	Processed atomic.Int64
	Dropped   atomic.Int64
	Errors    atomic.Int64
	Refs      atomic.Int64
}

type logTask struct {
	ctx       context.Context
	record    slog.Record
	timestamp time.Time
	startTime time.Time
	handler   slog.Handler // 携带 handler 引用
}

type worker struct {
	id          int
	workerGroup *WorkerGroup
	quitChan    chan struct{}
}

// 创建 WorkerGroup（单例管理）
func NewWorkerGroup(config *AsyncGroupConfig) *WorkerGroup {
	wg := &WorkerGroup{
		config:       config,
		logQueue:     make(chan *logTask, config.Queue),
		workers:      make([]*worker, 0, config.Worker),
		shutdownChan: make(chan struct{}),
		doneChan:     make(chan struct{}, config.Worker),
		stats:        &WorkerStats{},
		refCount:     atomic.Int64{},
	}

	wg.Start()

	return wg
}

// 启动 worker 池
func (wg *WorkerGroup) Start() {
	wg.mu.Lock()
	defer wg.mu.Unlock()

	if wg.workerCount.Load() > 0 {
		return // 已经启动
	}

	for i := 0; i < wg.config.Worker; i++ {
		w := &worker{
			id:          i,
			workerGroup: wg,
			quitChan:    make(chan struct{}),
		}
		wg.workers = append(wg.workers, w)
		go w.run()
	}

	wg.workerCount.Store(int64(len(wg.workers)))
}

// 停止 worker 池
func (wg *WorkerGroup) Stop() error {
	wg.mu.Lock()
	defer wg.mu.Unlock()

	if wg.workerCount.Load() == 0 {
		return nil // 已经停止
	}

	// 发送关闭信号
	close(wg.shutdownChan)

	// 等待所有 worker 完成
	stimeout := wg.config.ShutdownTimeout
	if stimeout <= 0 {
		stimeout = time.Second
	}
	timeout := time.After(stimeout)
	completed := 0
	for completed < wg.config.Worker {
		select {
		case <-wg.doneChan:
			completed++
		case <-timeout:
			return fmt.Errorf("worker group shutdown timeout after %v", wg.config.ShutdownTimeout)
		}
	}

	wg.workerCount.Store(0)
	return nil
}

// 增加引用（共享机制）
func (wg *WorkerGroup) Ref() {
	wg.refCount.Add(1)
	wg.stats.Refs.Add(1)
}

// 释放引用
func (wg *WorkerGroup) Unref() {
	count := wg.refCount.Add(-1)
	if count <= 0 {
		wg.Stop()
	}
}

// 提交日志任务（由 AsyncHandler 调用）
func (wg *WorkerGroup) Submit(task *logTask) error {
	switch wg.config.FullStrategy {
	case "drop_new":
		select {
		case wg.logQueue <- task:
			wg.stats.Queued.Add(1)
			return nil
		default:
			wg.stats.Dropped.Add(1)
			return nil
		}

	case "block_wait":
		wg.logQueue <- task
		wg.stats.Queued.Add(1)
		return nil

	case "drop_old":
		select {
		case wg.logQueue <- task:
			wg.stats.Queued.Add(1)
		default:
			select {
			case <-wg.logQueue:
				wg.stats.Dropped.Add(1)
				wg.logQueue <- task
			default:
			}
		}
		return nil
	}
	return nil
}

// 获取统计信息
func (wg *WorkerGroup) GetStats() *WorkerStats {
	return wg.stats
}

// Worker 运行逻辑
func (w *worker) run() {
	defer close(w.quitChan)
	defer func() {
		w.workerGroup.doneChan <- struct{}{}
	}()

	for {
		select {
		case task, ok := <-w.workerGroup.logQueue:
			if !ok {
				return
			}
			w.processTask(task)

		case <-w.workerGroup.shutdownChan:
			return
		}
	}
}

func (w *worker) processTask(task *logTask) {
	task.startTime = time.Now()

	// time.Sleep(time.Second) // TEST

	if err := task.handler.Handle(task.ctx, task.record); err != nil {
		w.workerGroup.stats.Errors.Add(1)
	} else {
		w.workerGroup.stats.Processed.Add(1)
	}
	// 这里需要 AsyncHandler 提供的原始 handler
	// 通过 task 携带或者 workerGroup 维护一个 handler 列表
}

// 2. WorkerGroupManager（管理多个 WorkerGroup）
// WorkerGroupManager 管理 WorkerGroup 的单例
type WorkerGroupManager struct {
	groups map[string]*WorkerGroup
	mu     sync.RWMutex
}

var globalManager = &WorkerGroupManager{
	groups: make(map[string]*WorkerGroup),
}

// 获取或创建 WorkerGroup
func GetWorkerGroup(name string, config *AsyncGroupConfig) *WorkerGroup {
	globalManager.mu.Lock()
	defer globalManager.mu.Unlock()

	if wg, exists := globalManager.groups[name]; exists {
		wg.Ref()
		return wg
	}

	wg := NewWorkerGroup(config)
	wg.Ref()
	globalManager.groups[name] = wg

	return wg
}

// 释放 WorkerGroup 引用
func ReleaseWorkerGroup(name string) {
	globalManager.mu.Lock()
	defer globalManager.mu.Unlock()

	if wg, exists := globalManager.groups[name]; exists {
		wg.Unref()
		if wg.stats.Refs.Load() <= 0 {
			delete(globalManager.groups, name)
		}
	}
}
