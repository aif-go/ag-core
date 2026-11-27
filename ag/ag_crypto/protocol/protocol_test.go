package protocol

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		label       string
		input       string
		wantOK      bool
		wantErr     bool
		wantName    string
		wantKey     string
		wantPayload string
	}{
		{"legacy", "{cipher}MTIz", true, false, "", "", "MTIz"},
		{"named", "{cipher}{name:base64}MTIz", true, false, "base64", "", "MTIz"},
		{"named+key", "{cipher}{name:sm4}{key:v2}AbCd", true, false, "sm4", "v2", "AbCd"},
		{"name uppercase", "{cipher}{name:SM4}{key:v2}AbCd", true, false, "sm4", "v2", "AbCd"},
		{"key before name", "{cipher}{key:v2}{name:sm4}AbCd", true, false, "sm4", "v2", "AbCd"},
		{"payload with brace", "{cipher}{name:sm4}abc}def", true, false, "sm4", "", "abc}def"},
		{"unknown header is payload", "{cipher}{foo:bar}MTIz", true, false, "", "", "{foo:bar}MTIz"},
		{"json payload", "{cipher}{\"k\":\"v\"}", true, false, "", "", "{\"k\":\"v\"}"},
		{"not cipher", "plain", false, false, "", "", ""},
		{"uppercase prefix", "{CIPHER}MTIz", false, false, "", "", ""},
		{"empty payload", "{cipher}", true, true, "", "", ""},
		{"empty payload with name", "{cipher}{name:sm4}", true, true, "", "", ""},
		{"key without name", "{cipher}{key:v2}AbCd", true, true, "", "", ""},
		{"empty name", "{cipher}{name:}MTIz", true, true, "", "", ""},
		{"empty key", "{cipher}{name:sm4}{key:}AbCd", true, true, "", "", ""},
		{"duplicate name", "{cipher}{name:sm4}{name:aes}AbCd", true, true, "", "", ""},
		{"unterminated", "{cipher}{name:sm4", true, true, "", "", ""},
		{"invalid name", "{cipher}{name:s m4}AbCd", true, true, "", "", ""},
		{"long name", "{cipher}{name:" + strings.Repeat("a", maxNameLen+1) + "}AbCd", true, true, "", "", ""},
	}

	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			ref, ok, err := Parse(c.input)
			if ok != c.wantOK {
				t.Fatalf("ok=%v, want %v (err=%v)", ok, c.wantOK, err)
			}
			if c.wantErr {
				if err == nil {
					t.Fatalf("want error, got ref=%+v", ref)
				}
				t.Logf("why=%s | err=%v", c.label, err)
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if ref.Name != c.wantName || ref.KeyID != c.wantKey || ref.Payload != c.wantPayload {
				t.Fatalf("ref=%+v, want name=%q key=%q payload=%q",
					ref, c.wantName, c.wantKey, c.wantPayload)
			}
		})
	}
}

func TestFormatRoundTrip(t *testing.T) {
	cases := []Ref{
		{Name: "", KeyID: "", Payload: "MTIz"},
		{Name: "base64", KeyID: "", Payload: "MTIz"},
		{Name: "sm4", KeyID: "v2", Payload: "AbCd"},
	}
	for _, c := range cases {
		s, err := Format(c.Name, c.KeyID, c.Payload)
		if err != nil {
			t.Fatalf("Format(%+v): %v", c, err)
		}
		got, ok, err := Parse(s)
		if err != nil || !ok {
			t.Fatalf("Parse(%q): ok=%v err=%v", s, ok, err)
		}
		if got != c {
			t.Fatalf("round-trip=%+v, want %+v", got, c)
		}
	}
}

func TestFormatRejectsKeyWithoutName(t *testing.T) {
	if _, err := Format("", "v2", "AbCd"); err == nil {
		t.Fatal("key without name should fail")
	}
}

// TestParseErrorDoesNotLeakCiphertext 保证解析错误只给格式提示、不回显密文，
// 避免错误被 slog 落盘时把密文片段带进日志。
func TestParseErrorDoesNotLeakCiphertext(t *testing.T) {
	const secret = "SECRET_CIPHERTEXT_123"
	cases := []string{
		"{cipher}{name:sm4}{" + secret, // 未闭合 header
	}
	for _, s := range cases {
		_, ok, err := Parse(s)
		if !ok || err == nil {
			t.Fatalf("Parse(%q): ok=%v err=%v, want ok=true err!=nil", s, ok, err)
		}
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("error leaks ciphertext: %v", err)
		}
	}
}

// TestFormatRejectsUnparseablePayload 保证 Format 不会产出 Parse 解析不回来的密文。
func TestFormatRejectsUnparseablePayload(t *testing.T) {
	for _, payload := range []string{
		"{name:x}",
		"{KEY:v1}data",
		"{abc",
	} {
		if _, err := Format("", "", payload); err == nil {
			t.Fatalf("Format(payload=%q) should fail", payload)
		}
	}
}

// TestFormatParseRoundTripWithBracePayload 验证以 { 开头的合法 payload 能原样往返，
// 包括 header 键带空格的写法（不会被 Parse 误认成 header）。
func TestFormatParseRoundTripWithBracePayload(t *testing.T) {
	for _, payload := range []string{
		`{"name":"x"}`,
		"{ name :x}rest",
		"{name :x}rest",
		"{foo:bar}baz",
	} {
		s, err := Format("", "", payload)
		if err != nil {
			t.Fatalf("Format(%q) should succeed: %v", payload, err)
		}
		got, ok, err := Parse(s)
		if err != nil || !ok {
			t.Fatalf("Parse(%q): ok=%v err=%v", s, ok, err)
		}
		if got.Name != "" || got.KeyID != "" || got.Payload != payload {
			t.Fatalf("round-trip=%+v, want payload=%q", got, payload)
		}
	}
}

// TestParseKeyWithoutNameMessage 钉住 {key:...} 缺 name 时的报错优先级。
func TestParseKeyWithoutNameMessage(t *testing.T) {
	_, ok, err := Parse("{cipher}{key:v2}")
	if !ok || err == nil {
		t.Fatalf("ok=%v err=%v, want ok=true err!=nil", ok, err)
	}
	if !strings.Contains(err.Error(), "requires {name:") {
		t.Fatalf("err=%v, want key-requires-name", err)
	}
}

func FuzzParse(f *testing.F) {
	f.Add("{cipher}MTIz")
	f.Add("{cipher}{name:sm4}{key:v2}AbCd")
	f.Fuzz(func(t *testing.T, s string) {
		_, _, _ = Parse(s)
	})
}
