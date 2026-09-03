package agonet

// 核心功能正确性测试（默认构建、绿测）：eventloop 主循环 + engine 生命周期。
// 覆盖：open/read/close 语义、run() 事件分发、Execute、engine shutdown 标记。

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aif-go/ag-core/contribute/agonet/pkg/aerrors"
)

// trackHandler 记录事件调用顺序的 EventHandler。
type trackHandler struct {
	opened   []*conn
	traffic  []*conn
	closed   []*conn
	closeErr []error
}

func (h *trackHandler) OnBoot(Engine) Action        { return None }
func (h *trackHandler) OnShutdown(Engine)           {}
func (h *trackHandler) OnOpen(c Conn) ([]byte, Action) {
	h.opened = append(h.opened, c.(*conn))
	return nil, None
}
func (h *trackHandler) OnTraffic(c Conn) Action {
	h.traffic = append(h.traffic, c.(*conn))
	return None
}
func (h *trackHandler) OnClose(c Conn, err error) Action {
	h.closed = append(h.closed, c.(*conn))
	h.closeErr = append(h.closeErr, err)
	return None
}

// newTestEventLoop 构造最小 eventloop（不启动 run goroutine）。
func newTestEventLoop(h EventHandler) *eventloop {
	ctx, cancel := context.WithCancel(context.Background())
	eng := &engine{
		opts:         &Options{},
		eventHandler: h,
		turnOff:      cancel,
	}
	_ = ctx
	return &eventloop{
		ch:           make(chan any, 8),
		idx:          0,
		eng:          eng,
		connections:  make(map[*conn]struct{}),
		eventHandler: h,
	}
}

func TestEventLoop_Open_RegistersAndCallsHandler(t *testing.T) {
	h := &trackHandler{}
	el := newTestEventLoop(h)
	c := newStreamConn(el, mockNetConn{}, nil)

	if err := el.open(&openConn{c: c}); err != nil {
		t.Fatalf("open err: %v", err)
	}
	if len(h.opened) != 1 {
		t.Fatalf("OnOpen calls=%d, expect 1", len(h.opened))
	}
	if _, ok := el.connections[c]; !ok {
		t.Fatal("conn not registered in eventloop")
	}
	if el.countConn() != 1 {
		t.Fatalf("connCount=%d, expect 1", el.countConn())
	}
}

func TestEventLoop_Read_IgnoresStaleConn(t *testing.T) {
	h := &trackHandler{}
	el := newTestEventLoop(h)
	c := newStreamConn(el, mockNetConn{}, nil)

	// 未注册连接：read 直接忽略，不触发 OnTraffic。
	if err := el.read(c); err != nil {
		t.Fatalf("read stale err: %v", err)
	}
	if len(h.traffic) != 0 {
		t.Fatalf("stale conn must not trigger OnTraffic, got %d", len(h.traffic))
	}
}

func TestEventLoop_Read_ConsumesBuffer(t *testing.T) {
	h := &trackHandler{}
	el := newTestEventLoop(h)
	c := newStreamConn(el, mockNetConn{}, nil)
	el.connections[c] = struct{}{}
	_, _ = c.buffer.Write([]byte("hello"))

	if err := el.read(c); err != nil {
		t.Fatalf("read err: %v", err)
	}
	if len(h.traffic) != 1 {
		t.Fatalf("OnTraffic calls=%d, expect 1", len(h.traffic))
	}
	// read 后 buffer 被清空（数据已转移到 inboundBuffer 或业务消费）
	if c.buffer.Len() != 0 {
		t.Fatalf("buffer not reset after read, len=%d", c.buffer.Len())
	}
}

func TestEventLoop_Close_RemovesAndReleases(t *testing.T) {
	h := &trackHandler{}
	el := newTestEventLoop(h)
	c := newStreamConn(el, mockNetConn{}, nil)
	el.connections[c] = struct{}{}
	el.incConn(1)

	if err := el.close(c, nil); err != nil {
		t.Fatalf("close err: %v", err)
	}
	if len(h.closed) != 1 {
		t.Fatalf("OnClose calls=%d, expect 1", len(h.closed))
	}
	if _, ok := el.connections[c]; ok {
		t.Fatal("conn still registered after close")
	}
	if el.countConn() != 0 {
		t.Fatalf("connCount=%d, expect 0", el.countConn())
	}
	if c.buffer != nil {
		t.Fatal("buffer not released after close")
	}
}

func TestEventLoop_Close_StaleIsNoop(t *testing.T) {
	h := &trackHandler{}
	el := newTestEventLoop(h)
	c := newStreamConn(el, mockNetConn{}, nil)

	// 未注册：close 幂等，不触发 OnClose。
	if err := el.close(c, nil); err != nil {
		t.Fatalf("close stale err: %v", err)
	}
	if len(h.closed) != 0 {
		t.Fatalf("stale close must not trigger OnClose, got %d", len(h.closed))
	}
}

func TestEventLoop_Execute_WhenShutdown(t *testing.T) {
	h := &trackHandler{}
	el := newTestEventLoop(h)
	el.eng.inShutdown.Store(true)

	if err := el.Execute(context.Background(), RunnableFunc(func(ctx context.Context) error { return nil })); !errors.Is(err, aerrors.ErrEngineInShutdown) {
		t.Fatalf("Execute after shutdown err=%v, expect ErrEngineInShutdown", err)
	}
}

func TestEventLoop_Execute_NilRunnable(t *testing.T) {
	h := &trackHandler{}
	el := newTestEventLoop(h)
	if err := el.Execute(context.Background(), nil); !errors.Is(err, aerrors.ErrNilRunnable) {
		t.Fatalf("Execute(nil) err=%v, expect ErrNilRunnable", err)
	}
}

func TestEventLoop_Run_DispatchAndShutdown(t *testing.T) {
	h := &trackHandler{}
	el := newTestEventLoop(h)

	// 启动 run goroutine，喂入各种事件，最后 ErrEngineShutdown 退出。
	go func() {
		_ = el.run()
	}()

	c := newStreamConn(el, mockNetConn{}, nil)
	el.ch <- &openConn{c: c}

	done := make(chan struct{})
	el.ch <- func() error {
		close(done)
		return nil
	}
	<-done // Execute 类型事件已处理

	el.ch <- aerrors.ErrEngineShutdown
	// run() 退出后：engine shutdown 被触发（turnOff 被调用）。
	// 等待 run goroutine 的 defer 完成（beingShutdown 置位）——轮询而非固定 sleep。
	deadline := time.Now().Add(2 * time.Second)
	for !el.eng.beingShutdown.Load() {
		if time.Now().After(deadline) {
			t.Fatal("engine beingShutdown not set after run exits")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestEngine_IsShutdown(t *testing.T) {
	eng := &engine{inShutdown: atomic.Bool{}}
	if eng.isShutdown() {
		t.Fatal("fresh engine must not be shutdown")
	}
	eng.inShutdown.Store(true)
	if !eng.isShutdown() {
		t.Fatal("engine must be shutdown after store")
	}
}

func TestEngine_Shutdown_MarksFlags(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	eng := &engine{turnOff: cancel, beingShutdown: atomic.Bool{}}
	eng.shutdown(nil)
	if !eng.beingShutdown.Load() {
		t.Fatal("beingShutdown not set")
	}
	select {
	case <-ctx.Done():
	default:
		t.Fatal("turnOff (context cancel) not invoked")
	}
}
