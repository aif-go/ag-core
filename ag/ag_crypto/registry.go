package ag_crypto

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/aif-go/ag-core/ag/ag_crypto/protocol"
)

type regKey struct {
	name  string
	keyID string
}

// EncryptorRegistry 按 (name, keyid) 管理加密器实例。
// 约定：启动期注册，运行期只读；内部用 RWMutex 保证并发安全。
type EncryptorRegistry struct {
	mu    sync.RWMutex
	byKey map[regKey]ITextEncryptor
	def   ITextEncryptor
}

func NewEncryptorRegistry() *EncryptorRegistry {
	return &EncryptorRegistry{byKey: make(map[regKey]ITextEncryptor)}
}

// Register 注册一个加密器实例；keyID 为空表示该 name 的默认实例。
func (r *EncryptorRegistry) Register(name, keyID string, enc ITextEncryptor) error {
	name = strings.ToLower(strings.TrimSpace(name))
	keyID = strings.TrimSpace(keyID)
	if err := protocol.ValidateName(name); err != nil {
		return err
	}
	if keyID != "" {
		if err := protocol.ValidateKeyID(keyID); err != nil {
			return err
		}
	}
	if enc == nil {
		return errors.New("nil encryptor")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	k := regKey{name: name, keyID: keyID}
	if _, ok := r.byKey[k]; ok {
		return fmt.Errorf("encryptor %q/%q already registered", name, keyID)
	}
	r.byKey[k] = enc
	return nil
}

// Resolve 按 (name, keyid) 取加密器。
// name 为空表示默认加密器；keyID 为空表示该 name 的默认实例。
func (r *EncryptorRegistry) Resolve(name, keyID string) (ITextEncryptor, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	keyID = strings.TrimSpace(keyID)

	r.mu.RLock()
	defer r.mu.RUnlock()

	if name == "" {
		if keyID != "" {
			return nil, fmt.Errorf("key id %q requires an encryptor name", keyID)
		}
		if r.def == nil {
			return nil, errors.New("no default encryptor registered")
		}
		return r.def, nil
	}

	if enc, ok := r.byKey[regKey{name: name, keyID: keyID}]; ok {
		return enc, nil
	}

	var keys []string
	for k := range r.byKey {
		if k.name == name {
			keys = append(keys, k.keyID)
		}
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("encryptor %q not registered", name)
	}
	sort.Strings(keys)
	return nil, fmt.Errorf("encryptor %q has no key id %q (available: %v)", name, keyID, keys)
}

// SetDefault 设置 name 为空时使用的默认加密器。
// 建议在启动期调用；运行期调用时，已在执行的解密不受影响，之后开始的解密使用新值。
func (r *EncryptorRegistry) SetDefault(enc ITextEncryptor) error {
	if enc == nil {
		return errors.New("nil encryptor")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.def = enc
	return nil
}

// Default 返回默认加密器；未设置时为 nil。
func (r *EncryptorRegistry) Default() ITextEncryptor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.def
}
