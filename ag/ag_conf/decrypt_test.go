package ag_conf

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aif-go/ag-core/ag/ag_crypto"
)

func newTestSource(name string, source map[string]any) *MapPropertySource {
	return &MapPropertySource{
		NamedPropertySource: NamedPropertySource{Name: name},
		Source:              source,
	}
}

func TestBuildDecryptForPropertySource(t *testing.T) {
	cases := []struct {
		label   string
		input   map[string]any
		want    map[string]any
		wantErr string
	}{
		{
			"legacy base64",
			map[string]any{"pwd": "{cipher}MTIzNDU2"},
			map[string]any{"pwd": "123456"},
			"",
		},
		{
			"named base64",
			map[string]any{"pwd": "{cipher}{name:base64}MTIzNDU2"},
			map[string]any{"pwd": "123456"},
			"",
		},
		{
			"unknown encryptor",
			map[string]any{"pwd": "{cipher}{name:sm4}AAAA"},
			nil,
			"sm4",
		},
		{
			"key without name",
			map[string]any{"pwd": "{cipher}{key:v2}AAAA"},
			nil,
			"requires",
		},
		{
			"plain value",
			map[string]any{"pwd": "plain"},
			map[string]any{},
			"",
		},
	}

	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			ps := newTestSource("[LOCAL]-test", c.input)
			got, err := BuildDecryptForPropertySource(ps)
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("err=%v, want contains %q", err, c.wantErr)
				}
				t.Logf("why=%s | err=%v", c.label, err)
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if !reflect.DeepEqual(got.Source, c.want) {
				t.Fatalf("source=%v, want %v", got.Source, c.want)
			}
		})
	}
}

type fakeEncryptor struct{}

func (fakeEncryptor) Name() string                     { return "fake" }
func (fakeEncryptor) Encrypt(s string) (string, error) { return s, nil }
func (fakeEncryptor) Decrypt(s string) (string, error) { return s, nil }

// TestBuildDecryptHonorsLegacyDefaultVar 验证下游直接赋值 DefaultEncryptorPrimary
// （旧 API 的公开变量）仍然生效，不会被静默忽略。
func TestBuildDecryptHonorsLegacyDefaultVar(t *testing.T) {
	old := ag_crypto.DefaultEncryptorPrimary
	ag_crypto.DefaultEncryptorPrimary = fakeEncryptor{}
	t.Cleanup(func() { ag_crypto.DefaultEncryptorPrimary = old })

	ps := newTestSource("[LOCAL]-legacy-var", map[string]any{"pwd": "{cipher}whatever"})
	got, err := BuildDecryptForPropertySource(ps)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Source["pwd"] != "whatever" {
		t.Fatalf("pwd=%v, want whatever（直接赋值 DefaultEncryptorPrimary 应生效）", got.Source["pwd"])
	}
}
