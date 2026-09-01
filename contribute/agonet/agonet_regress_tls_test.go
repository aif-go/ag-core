//go:build agonet_regress

package agonet

// 回归测试（white-box, package agonet）：E 根因组（TLS/TLCP 安全配置）。
// 对应跟踪清单：agonet-问题总报告与跟踪.md -> E1 / E3（commit 4d1491a4 快照）。
// TDD 红-绿语义：断言"正确行为"，缺陷未修复时【预期失败(红)】，修复后【通过(绿)】自动验证。
// 运行：go test -race -tags agonet_regress . -run 'TestE1|TestE3' -v

import (
	"net"
	"testing"
)

// TestE1_TlsTlcpClientShouldApplyTLS 缺陷复现
//
// 期望行为：Type=tls_tlcp + CliType=tls 时，客户端 TLS 类型应归一化为 TLS（非空），
//           DialContext 走 tls.Dial 而非明文 net.Dial。
// 当前缺陷：NewClient 守卫 `Type != TLSTYPE_TLS_TLCP` 使 WithAgClientTLSConfig 完全不执行，
//           CliTLSType()=="" -> DialContext 落入 default: net.Dial 明文。
func TestE1_TlsTlcpClientShouldApplyTLS(t *testing.T) {
	cfg := DefaultClientConfig()
	cfg.Config.Security.Type = TLSTYPE_TLS_TLCP
	cfg.Config.Security.CliType = TLSType_TLS // 显式 cliType：客户端应只走 TLS

	cli, err := NewClient(&BuiltinEventEngine{}, &cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	got := cli.(*client).opts.CliTLSType()
	if got != TLSType_TLS {
		t.Fatalf("FAIL(E1 未修复): CliTLSType()=%q, 期望 %q（TLS 配置被跳过 -> DialContext 走明文 net.Dial）",
			got, TLSType_TLS)
	}
	t.Logf("PASS(E1 已修复): CliTLSType()=%q, 客户端将走 TLS 而非明文", got)
}

// TestE3_TlsTypeWithoutConfigShouldError 缺陷复现
//
// 缺陷：tlsIfNeed（listener.go:81-98）仅在 TLSType!=NONE 且任一 cfg 非 nil 时才包装；
//       若 Type=tls 但 TLSConfig/TLCPConfig 均为 nil，则静默不包装 -> 明文监听，无告警。
//       单协议配置（如 Type=tls 只有 TLCPConfig）还会被 pa.NewListener 单边 nil 静默改判协议。
// 期望：Type 已设但对应 config 缺失 -> createListener 返回错误（ErrTLSConfigIsNil），不静默明文。
func TestE3_TlsTypeWithoutConfigShouldError(t *testing.T) {
	opts := &Options{TLSType: TLSType_TLS} // 声明 TLS 但无 TLSConfig/TLCPConfig

	l, err := createListener("tcp", "127.0.0.1:0", opts)
	if l != nil && l.ln != nil {
		defer l.ln.Close()
	}
	if err == nil {
		t.Fatalf("FAIL(E3 未修复): Type=tls 但 config 为 nil 时 createListener 静默成功(明文监听)；期望返回错误(ErrTLSConfigIsNil)")
	}
	if err != nil && err != net.ErrClosed {
		t.Logf("PASS(E3 已修复): Type 与 config 不一致时返回错误: %v", err)
	}
}
