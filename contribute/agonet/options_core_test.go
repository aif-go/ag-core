package agonet

// 核心功能正确性测试（默认构建、绿测）：Options getter fallback + Validate 校验。

import (
	"crypto/tls"
	"testing"

	"gitee.com/Trisia/gotlcp/tlcp"
)

func TestCliTLSType_Fallback(t *testing.T) {
	// CLI 未设 → fallback 到服务端 TLSType（复用用法）。
	o := &Options{TLSType: TLSType_TLS}
	if got := o.CliTLSType(); got != TLSType_TLS {
		t.Fatalf("CliTLSType()=%q, expect tls (fallback)", got)
	}
}

func TestCliTLSType_ExplicitWins(t *testing.T) {
	// CLI 显式设置 → 不 fallback。
	o := &Options{TLSType: TLSType_TLCP, CLI_TLSType: TLSType_TLS}
	if got := o.CliTLSType(); got != TLSType_TLS {
		t.Fatalf("CliTLSType()=%q, expect tls (explicit CLI wins)", got)
	}
}

func TestCliTLSConfig_Fallback(t *testing.T) {
	// CLI config 未设 → fallback 服务端 TLSConfig（复用证书）。
	serverCfg := &tls.Config{}
	o := &Options{TLSConfig: serverCfg}
	if got := o.CliTLSConfig(); got != serverCfg {
		t.Fatalf("CliTLSConfig() must fallback to server TLSConfig")
	}
}

func TestCliTLCPConfig_Fallback(t *testing.T) {
	serverCfg := &tlcp.Config{}
	o := &Options{TLCPConfig: serverCfg}
	if got := o.CliTLCPConfig(); got != serverCfg {
		t.Fatalf("CliTLCPConfig() must fallback to server TLCPConfig")
	}
}

func TestValidateServer_UnsetNone_OK(t *testing.T) {
	for _, tt := range []TLSType{TLSType_UNSET, TLSType_NONE, ""} {
		o := &Options{TLSType: tt}
		if err := o.ValidateServer(); err != nil {
			t.Fatalf("ValidateServer(%q) err=%v, expect nil (明文合法)", tt, err)
		}
	}
}

func TestValidateServer_TLS_MissingConfig_Error(t *testing.T) {
	o := &Options{TLSType: TLSType_TLS}
	if err := o.ValidateServer(); err == nil {
		t.Fatal("ValidateServer(tls, no config) must error")
	}
}

func TestValidateServer_TLS_WithConfig_OK(t *testing.T) {
	o := &Options{TLSType: TLSType_TLS, TLSConfig: &tls.Config{}}
	if err := o.ValidateServer(); err != nil {
		t.Fatalf("ValidateServer err=%v, expect nil", err)
	}
}

func TestValidateServer_TLCP_MissingConfig_Error(t *testing.T) {
	o := &Options{TLSType: TLSType_TLCP}
	if err := o.ValidateServer(); err == nil {
		t.Fatal("ValidateServer(tlcp, no config) must error")
	}
}

func TestValidateServer_TlsTlcp_RequiresBoth(t *testing.T) {
	// 只配一个 config → 报错。
	if err := (&Options{TLSType: TLSTYPE_TLS_TLCP, TLSConfig: &tls.Config{}}).ValidateServer(); err == nil {
		t.Fatal("tls_tlcp with only TLSConfig must error")
	}
	if err := (&Options{TLSType: TLSTYPE_TLS_TLCP, TLCPConfig: &tlcp.Config{}}).ValidateServer(); err == nil {
		t.Fatal("tls_tlcp with only TLCPConfig must error")
	}
	// 两个都配 → 通过。
	o := &Options{TLSType: TLSTYPE_TLS_TLCP, TLSConfig: &tls.Config{}, TLCPConfig: &tlcp.Config{}}
	if err := o.ValidateServer(); err != nil {
		t.Fatalf("tls_tlcp with both configs err=%v, expect nil", err)
	}
}

func TestValidateServer_UnknownType_Error(t *testing.T) {
	o := &Options{TLSType: "quic"}
	if err := o.ValidateServer(); err == nil {
		t.Fatal("ValidateServer(unknown) must error")
	}
}

func TestValidateClient_Unset_PlaintextDefault(t *testing.T) {
	// 客户端未声明安全类型：明文合法默认。
	o := &Options{}
	if err := o.ValidateClient(); err != nil {
		t.Fatalf("ValidateClient(empty) err=%v, expect nil (明文默认)", err)
	}
}

func TestValidateClient_ReuseServerTLSConfig(t *testing.T) {
	// 复用用法：CLI 未设，借服务端 TLSType+TLSConfig → 通过（不误判）。
	o := &Options{TLSType: TLSType_TLS, TLSConfig: &tls.Config{}}
	if err := o.ValidateClient(); err != nil {
		t.Fatalf("ValidateClient(reuse server cfg) err=%v, expect nil", err)
	}
}

func TestValidateClient_ExplicitTLS_MissingConfig_Error(t *testing.T) {
	o := &Options{CLI_TLSType: TLSType_TLS}
	if err := o.ValidateClient(); err == nil {
		t.Fatal("ValidateClient(CLI tls, no config) must error")
	}
}

func TestValidateClient_TlsTlcp_Rejected(t *testing.T) {
	// 手工 Options fallback 到 tls_tlcp → 拒绝（DialContext 无对应分支，静默明文）。
	o := &Options{TLSType: TLSTYPE_TLS_TLCP, TLSConfig: &tls.Config{}, TLCPConfig: &tlcp.Config{}}
	if err := o.ValidateClient(); err == nil {
		t.Fatal("ValidateClient(tls_tlcp fallback) must error")
	}
}

func TestValidateClient_ExplicitTLS_WithConfig_OK(t *testing.T) {
	o := &Options{CLI_TLSType: TLSType_TLS, CLI_TLSConfig: &tls.Config{}}
	if err := o.ValidateClient(); err != nil {
		t.Fatalf("ValidateClient(CLI tls + config) err=%v, expect nil", err)
	}
}
