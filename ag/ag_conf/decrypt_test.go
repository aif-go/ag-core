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

// TestBuildDecryptHonorsRegistryDefault 验证用 SetDefault 设置默认加密器后，
// legacy {cipher} 也会走新的默认值，而不只是注册表解析。
func TestBuildDecryptHonorsRegistryDefault(t *testing.T) {
	oldDef := ag_crypto.DefaultRegistry().Default()
	t.Cleanup(func() {
		_ = ag_crypto.DefaultRegistry().SetDefault(oldDef)
	})

	if err := ag_crypto.DefaultRegistry().SetDefault(fakeEncryptor{}); err != nil {
		t.Fatal(err)
	}
	ps := newTestSource("[LOCAL]-registry-default", map[string]any{"pwd": "{cipher}@@@"})
	got, err := BuildDecryptForPropertySource(ps)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Source["pwd"] != "@@@" {
		t.Fatalf("pwd=%v, want @@@（SetDefault 应影响 legacy {cipher}）", got.Source["pwd"])
	}
}
