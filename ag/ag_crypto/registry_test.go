package ag_crypto

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestRegistryRegisterAndResolve(t *testing.T) {
	r := NewEncryptorRegistry()
	if err := r.Register("base64", "", Base64Encryptor); err != nil {
		t.Fatal(err)
	}
	if err := r.Register("sm4", "v1", Base64Encryptor); err != nil {
		t.Fatal(err)
	}
	if err := r.Register("sm4", "v2", Base64Encryptor); err != nil {
		t.Fatal(err)
	}
	if err := r.SetDefault(Base64Encryptor); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name    string
		keyID   string
		wantErr string
	}{
		{"", "", ""},
		{"base64", "", ""},
		{"sm4", "v1", ""},
		{"sm4", "v2", ""},
		{"SM4", "v1", ""}, // name 大小写不敏感
		{"sm4", "", "no key id"},
		{"aes", "", "not registered"},
		{"sm4", "v9", "v9"},
		{"", "v2", "requires"},
	}
	for _, c := range cases {
		t.Run(c.name+"/"+c.keyID, func(t *testing.T) {
			_, err := r.Resolve(c.name, c.keyID)
			if c.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("err=%v, want contains %q", err, c.wantErr)
			}
		})
	}
}

func TestRegistryDuplicateAndInvalid(t *testing.T) {
	r := NewEncryptorRegistry()
	if err := r.Register("base64", "", Base64Encryptor); err != nil {
		t.Fatal(err)
	}
	if err := r.Register("base64", "", Base64Encryptor); err == nil {
		t.Fatal("duplicate registration should fail")
	}
	if err := r.Register("", "", Base64Encryptor); err == nil {
		t.Fatal("empty name should fail")
	}
	if err := r.Register("bad name", "", Base64Encryptor); err == nil {
		t.Fatal("invalid name should fail")
	}
	if err := r.Register("ok", "", nil); err == nil {
		t.Fatal("nil encryptor should fail")
	}
}

func TestRegistryConcurrent(t *testing.T) {
	r := NewEncryptorRegistry()
	for i := 0; i < 10; i++ {
		if err := r.Register(fmt.Sprintf("enc%d", i), "", Base64Encryptor); err != nil {
			t.Fatal(err)
		}
	}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _ = r.Resolve(fmt.Sprintf("enc%d", i%10), "")
		}(i)
	}
	wg.Wait()
}

// TestConcurrentDefaultAccess 验证 DefaultRegistry / GetEncrytorPrimary / SetDefault
// 三路并发访问不会产生数据竞争（配合 -race 运行）。
func TestConcurrentDefaultAccess(t *testing.T) {
	old := DefaultRegistry().Default()
	if old == nil {
		old = Base64Encryptor
	}
	t.Cleanup(func() { _ = DefaultRegistry().SetDefault(old) })

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			_ = DefaultRegistry()
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			_ = GetEncrytorPrimary()
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			_ = DefaultRegistry().SetDefault(old)
		}
	}()
	wg.Wait()
}
