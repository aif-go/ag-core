package async

import (
	"context"
	"log/slog"
	"time"
)

var (
	_ slog.Handler = (*AsyncHandler)(nil)
)

// AsyncHandler 包装层，只负责引用和委托
type AsyncHandler struct {
	workerGroup *WorkerGroup
	original    slog.Handler
}

// 创建 AsyncHandler
func NewAsyncHandler(original slog.Handler, groupName string, config *AsyncGroupConfig) slog.Handler {
	wg := GetWorkerGroup(groupName, config)

	return &AsyncHandler{
		workerGroup: wg,
		original:    original,
	}
}

// 关闭 AsyncHandler
func (h *AsyncHandler) Close() error {
	// 这里只是减少引用计数
	// 不直接调用 workerGroup.Stop()
	// Stop 由 WorkerGroupManager 管理
	return nil
}

// slog.Handler 接口实现
func (h *AsyncHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.original.Enabled(ctx, level)
}
func (h *AsyncHandler) Handle(ctx context.Context, r slog.Record) error {
	task := &logTask{
		ctx:       ctx,
		record:    r,
		timestamp: time.Now(),
		handler:   h.original,
	}

	// 委托给 WorkerGroup 处理
	return h.workerGroup.Submit(task)
}
func (h *AsyncHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &AsyncHandler{
		workerGroup: h.workerGroup,
		original:    h.original.WithAttrs(attrs),
	}
}
func (h *AsyncHandler) WithGroup(name string) slog.Handler {
	return &AsyncHandler{
		workerGroup: h.workerGroup,
		original:    h.original.WithGroup(name),
	}
}

// 获取统计信息
func (h *AsyncHandler) GetStats() *WorkerStats {
	return h.workerGroup.GetStats()
}



