package agonet

// 核心功能正确性测试（默认构建、绿测）：config 默认值 + parseProtoAddr。

import (
	"testing"
)

func TestDefaultConfig_NonZero(t *testing.T) {
	sc := DefaultServerConfig()
	if sc.Addr == "" {
		t.Fatal("DefaultServerConfig.Addr empty")
	}
	// 默认引擎配置 NumEventLoop=0（引擎自行确定），合法默认。
	_ = DefaultClientConfig()
}

func TestDefaultConfig_NoTLSByDefault(t *testing.T) {
	sc := DefaultServerConfig()
	if sc.Config.Security.Type != TLSType_NONE && sc.Config.Security.Type != TLSType_UNSET {
		t.Fatalf("default security type=%q, expect none/unset", sc.Config.Security.Type)
	}
}

func TestParseProtoAddr(t *testing.T) {
	proto, addr, err := parseProtoAddr("tcp://127.0.0.1:8080")
	if err != nil {
		t.Fatal(err)
	}
	if proto != "tcp" {
		t.Fatalf("proto=%q, expect tcp", proto)
	}
	if addr != "127.0.0.1:8080" {
		t.Fatalf("addr=%q, expect 127.0.0.1:8080", addr)
	}
}

func TestParseProtoAddr_NoScheme_Error(t *testing.T) {
	// 无 scheme：解析失败返回 ErrInvalidNetworkAddress（防御正确行为）。
	_, _, err := parseProtoAddr("127.0.0.1:8080")
	if err == nil {
		t.Fatal("parseProtoAddr without scheme must error")
	}
}
