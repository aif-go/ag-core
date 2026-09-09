package simple

// 核心功能正确性测试（默认构建、绿测）：pipeline 链语义。
// 覆盖：AddFirst/AddLast 插入顺序、FireRead 正向遍历、FireWrite 反向遍历、
//       双向 handler（DuplexHandler）同时参与入站/出站。

import (
	"testing"
)

// recordInbound 记录被调用的入站 handler 顺序。
type recordInbound struct {
	name string
	log  *[]string
}

func (h *recordInbound) HandleRead(ctx InboundContext, message any) {
	*h.log = append(*h.log, "in:"+h.name)
}

// recordOutbound 记录被调用的出站 handler 顺序。
type recordOutbound struct {
	name string
	log  *[]string
}

func (h *recordOutbound) HandleWrite(ctx OutboundContext, message any) {
	*h.log = append(*h.log, "out:"+h.name)
}

// recordDuplex 双向：入站/出站都记录。
type recordDuplex struct {
	recordInbound
	recordOutbound
}

func newDuplex(name string, log *[]string) *recordDuplex {
	return &recordDuplex{
		recordInbound:  recordInbound{name: name, log: log},
		recordOutbound: recordOutbound{name: name, log: log},
	}
}

func TestPipeline_AddLast_KeepsOrder(t *testing.T) {
	p := NewPipeline()
	var log []string
	p.AddLast(&recordInbound{name: "a", log: &log}, &recordInbound{name: "b", log: &log})

	p.FireChannelRead([]byte("m"))
	if len(log) != 1 || log[0] != "in:a" {
		t.Fatalf("FireRead must hit first inbound only, log=%v", log)
	}
	// 只有首个 inbound handler 被调用（后续靠 handler 显式 FireRead 传播）
}

func TestPipeline_AddFirst_ReversesOrder(t *testing.T) {
	p := NewPipeline()
	var log []string
	p.AddFirst(&recordInbound{name: "a", log: &log})
	p.AddFirst(&recordInbound{name: "b", log: &log})

	p.FireChannelRead([]byte("m"))
	if len(log) != 1 || log[0] != "in:b" {
		t.Fatalf("AddFirst: last added must be first hit, log=%v", log)
	}
}

func TestPipeline_AddLast_ChainPropagation(t *testing.T) {
	// handler 显式 FireRead 传播：a→b→c 依次执行。
	p := NewPipeline()
	var log []string
	propagate := func(name string) InboundHandler {
		return &propagateInbound{name: name, log: &log}
	}
	p.AddLast(propagate("a"), propagate("b"), propagate("c"))

	p.FireChannelRead([]byte("m"))
	want := []string{"in:a", "in:b", "in:c"}
	if len(log) != len(want) {
		t.Fatalf("chain log=%v, want %v", log, want)
	}
	for i := range want {
		if log[i] != want[i] {
			t.Fatalf("chain log=%v, want %v", log, want)
		}
	}
}

// propagateInbound 入站时记录并继续传播。
type propagateInbound struct {
	name string
	log  *[]string
}

func (h *propagateInbound) HandleRead(ctx InboundContext, message any) {
	*h.log = append(*h.log, "in:"+h.name)
	ctx.FireRead(message)
}

func TestPipeline_FireWrite_ReverseOrder(t *testing.T) {
	// 出站从 tail 反向：最后一个 AddLast 的 outbound 最先执行。
	// absorb 放 head 侧（AddFirst），传播链 a→b→absorb 终止，避免到达 headHandler 的 Write1。
	p := NewPipeline()
	var log []string
	prop := func(name string) OutboundHandler {
		return &propagateOutbound{name: name, log: &log}
	}
	p.AddFirst(&absorbOutbound{})
	p.AddLast(prop("a"), prop("b"))

	p.FireChannelWrite([]byte("m"))
	want := []string{"out:b", "out:a"} // 反向：b 先于 a
	if len(log) != len(want) {
		t.Fatalf("outbound log=%v, want %v", log, want)
	}
	for i := range want {
		if log[i] != want[i] {
			t.Fatalf("outbound log=%v, want %v", log, want)
		}
	}
}

// absorbOutbound 吸收消息不传播：终止出站链。
type absorbOutbound struct{}

func (h *absorbOutbound) HandleWrite(ctx OutboundContext, message any) {}

// propagateOutbound 出站时记录并继续传播（向前）。
type propagateOutbound struct {
	name string
	log  *[]string
}

func (h *propagateOutbound) HandleWrite(ctx OutboundContext, message any) {
	*h.log = append(*h.log, "out:"+h.name)
	ctx.FireWrite(message)
}

func TestPipeline_DuplexHandler_BothDirections(t *testing.T) {
	p := NewPipeline()
	var log []string
	p.AddFirst(&absorbOutbound{})
	p.AddLast(newDuplex("d", &log))

	p.FireChannelRead([]byte("m"))
	if len(log) != 1 || log[0] != "in:d" {
		t.Fatalf("duplex inbound log=%v", log)
	}
	log = nil
	p.FireChannelWrite([]byte("m"))
	if len(log) != 1 || log[0] != "out:d" {
		t.Fatalf("duplex outbound log=%v", log)
	}
}

func TestPipeline_AddFirstAddLast_Mixed(t *testing.T) {
	// AddFirst(a), AddLast(b,c)：入站顺序 a→b→c（a 在 head 侧）。
	p := NewPipeline()
	var log []string
	prop := func(name string) InboundHandler { return &propagateInbound{name: name, log: &log} }
	p.AddFirst(prop("a"))
	p.AddLast(prop("b"), prop("c"))

	p.FireChannelRead([]byte("m"))
	want := []string{"in:a", "in:b", "in:c"}
	if len(log) != len(want) {
		t.Fatalf("mixed log=%v, want %v", log, want)
	}
	for i := range want {
		if log[i] != want[i] {
			t.Fatalf("mixed log=%v, want %v", log, want)
		}
	}
}

func TestPipeline_Size(t *testing.T) {
	p := NewPipeline().(*pipeline)
	if p.size != 2 {
		t.Fatalf("initial pipeline size=%d, expect 2 (head+tail)", p.size)
	}
	p.AddLast(&recordInbound{name: "a"})
	p.AddLast(&recordInbound{name: "b"})
	p.AddFirst(&recordInbound{name: "c"})
	if p.size != 5 {
		t.Fatalf("pipeline size=%d, expect 5", p.size)
	}
}
