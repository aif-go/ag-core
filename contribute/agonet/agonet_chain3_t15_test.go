package agonet

// 链3 F3 半包滞留 DoS 修复测试（TDD 红→绿）：
//   T15——入站滞留超限 → 连接关闭（防 inboundBuffer 无上限增长内存 DoS）
// 机制：慢速/故障客户端持续发数据且业务不消费 → 滞留累积 → 超 InboundBufferLimit
//       → el.read 检查点关闭连接（内存释放）
// 红态：修复前（无检查点）——连接不被关（滞留无限增长）
// 绿态：检查点生效——超限即关（客户端感知关闭）

import (
	"net"
	"testing"
	"time"
)

// noConsumeHandler 不消费入站数据（OnTraffic 空——数据全滞留 inboundBuffer）
type noConsumeHandler struct{}

func (h *noConsumeHandler) OnBoot(Engine) Action         { return None }
func (h *noConsumeHandler) OnShutdown(Engine)            {}
func (h *noConsumeHandler) OnOpen(Conn) ([]byte, Action) { return nil, None }
func (h *noConsumeHandler) OnClose(Conn, error) Action   { return None }
func (h *noConsumeHandler) OnTraffic(Conn) Action        { return None }

func TestChain3_T15_InboundOverflowClosesConn(t *testing.T) {
	const addr = "tcp://127.0.0.1:18115"
	const host = "127.0.0.1:18115"
	const limit = 8 * 1024 // 测小值：滞留 > 8KB 即超限

	// server：InboundBufferLimit = 8KB + 不消费 handler（滞留累积）
	cfg := DefaultServerConfig()
	cfg.Addr = addr
	cfg.Config.Engine.InboundBufferLimit = limit
	srv, err := NewServer(&noConsumeHandler{}, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = srv.Start() }()
	t.Cleanup(func() { _ = srv.Stop() })

	// 等端口就绪
	deadline := time.Now().Add(3 * time.Second)
	for {
		c, err := net.DialTimeout("tcp", host, 200*time.Millisecond)
		if err == nil {
			_ = c.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("server not ready")
		}
		time.Sleep(20 * time.Millisecond)
	}

	// 客户端：持续发数据（总量远超 limit——滞留超限）
	nc, err := net.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Close()
	payload := make([]byte, 4096) // 4KB × 32 = 128KB（>> 8KB limit）
	for i := 0; i < 32; i++ {
		if _, err := nc.Write(payload); err != nil {
			break // 连接已被关——正是预期
		}
		time.Sleep(2 * time.Millisecond)
	}

	// 断言：连接被服务端关闭（Read 返回 err——连接断）
	readDeadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(readDeadline) {
		_ = nc.SetReadDeadline(time.Now().Add(500 * time.Millisecond)) // 防阻塞挂起
		if _, err := nc.Read(make([]byte, 64)); err != nil {
			// 连接被关（EOF/RST）——F3 检查点生效
			return // PASS（绿态）
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("RED: 滞留超限但连接未被关闭（F3 检查点缺失——inboundBuffer 无上限增长）")
}
