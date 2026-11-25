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
		{"old form", "{cipher:sm4}AbCd", true, true, "", "", ""},
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

func FuzzParse(f *testing.F) {
	f.Add("{cipher}MTIz")
	f.Add("{cipher}{name:sm4}{key:v2}AbCd")
	f.Add("{cipher:sm4}AbCd")
	f.Fuzz(func(t *testing.T, s string) {
		_, _, _ = Parse(s)
	})
}
