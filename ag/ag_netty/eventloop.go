package ag_netty

import (
	"context"
	"github.com/panjf2000/gnet"
	"net"
	"sync"
	"time"
)

// EventLoop 事件循环
type EventLoop struct {
	server    *gnet.Server
	handler   *gnetServerHandler
	taskQueue chan func()
	quit      chan struct{}
	wg        sync.WaitGroup
	connMap   sync.Map // 存储连接的映射
	initFunc  func(ch *Channel)
}

type gnetServerHandler struct {
	el *EventLoop
}

// NewEventLoop 创建新事件循环
func NewEventLoop(initFunc func(ch *Channel)) (*EventLoop, error) {
	el := &EventLoop{
		taskQueue: make(chan func(), 1024),
		quit:      make(chan struct{}),
		initFunc:  initFunc,
	}

	el.handler = &gnetServerHandler{el: el}

	// 启动任务处理协程
	el.wg.Add(1)
	go el.runTaskLoop()

	return el, nil
}

func (h *gnetServerHandler) OnInitComplete(server gnet.Server) (action gnet.Action) {
	h.el.server = &server
	return gnet.None
}

func (h *gnetServerHandler) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	conn := NewGnetConnAdapter(c)
	channel := NewChannel(conn, h.el)

	h.el.connMap.Store(c, channel)

	if h.el.initFunc != nil {
		h.el.initFunc(channel)
	}

	channel.Pipeline.FireActive()

	return nil, gnet.None
}

func (h *gnetServerHandler) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	if ch, ok := h.el.connMap.LoadAndDelete(c); ok {
		channel := ch.(*Channel)
		channel.Close()
	}
	return gnet.None
}

func (h *gnetServerHandler) React(packet []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	if ch, ok := h.el.connMap.Load(c); ok {
		channel := ch.(*Channel)

		data := make([]byte, len(packet))
		copy(data, packet)

		channel.Pipeline.FireRead(data)
	}
	return nil, gnet.None
}

func (h *gnetServerHandler) PreWrite(c gnet.Conn) {
	// No-op
}

func (h *gnetServerHandler) AfterWrite(c gnet.Conn, b []byte) {
	// No-op
}

func (h *gnetServerHandler) Tick() (delay time.Duration, action gnet.Action) {
	return time.Hour, gnet.None // Long delay as we don't use ticking
}

func (h *gnetServerHandler) OnShutdown(server gnet.Server) {
	// No-op
}

// runTaskLoop 运行任务处理循环
func (el *EventLoop) runTaskLoop() {
	defer el.wg.Done()

	for {
		select {
		case task := <-el.taskQueue:
			task()
		case <-el.quit:
			return
		}
	}
}

// IsShutdown 检查是否已关闭
func (el *EventLoop) IsShutdown() bool {
	select {
	case <-el.quit:
		return true
	default:
		return false
	}
}

// Post 投递任务到事件循环
func (el *EventLoop) Post(task func()) {
	select {
	case el.taskQueue <- task:
	case <-el.quit:
	}
}

// Schedule 调度延迟任务
func (el *EventLoop) Schedule(delay time.Duration, task func()) {
	time.AfterFunc(delay, func() {
		el.Post(task)
	})
}

// Run 运行事件循环
func (el *EventLoop) Run(listener net.Listener, numEventLoops int) error {
	network := "tcp"
	addr := listener.Addr().String()

	listener.Close()

	return gnet.Serve(el.handler, network+"://"+addr, gnet.WithNumEventLoop(numEventLoops))
}

// Shutdown 关闭事件循环
func (el *EventLoop) Shutdown() {
	close(el.quit)
	el.wg.Wait()

	if el.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		gnet.Stop(ctx, el.server.Addr.String())
	}

	el.connMap.Range(func(key, value interface{}) bool {
		if ch, ok := value.(*Channel); ok {
			ch.Close()
		}
		return true
	})
}
