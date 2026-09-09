package agonet

// 回归测试（white-box, package agonet）：E 根因组（TLS/TLCP 安全配置）。
// 对应跟踪清单：agonet-问题总报告与跟踪.md -> E1 / E3（commit 4d1491a4 快照）。
// 修复后已转绿，移出 agonet_regress tag，纳入默认构建。
// 运行：go test ./contribute/agonet/ -run 'TestE1|TestE3' -v

import (
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
// 期望：配置自洽校验归属 Options —— ValidateServer 在构造期返回错误（ErrTLSConfigIsNil），不静默明文。
func TestE3_TlsTypeWithoutConfigShouldError(t *testing.T) {
	opts := &Options{TLSType: TLSType_TLS} // 声明 TLS 但无 TLSConfig/TLCPConfig

	err := opts.ValidateServer()
	if err == nil {
		t.Fatalf("FAIL(E3 未修复): Type=tls 但 config 为 nil 时 ValidateServer 通过；期望返回错误(ErrTLSConfigIsNil)")
	}
	t.Logf("PASS(E3 已修复): Type 与 config 不一致时构造期返回错误: %v", err)
}
