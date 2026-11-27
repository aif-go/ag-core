package ag_crypto

import "sync"

// registryMu 保护 defaultRegistry 指针。当前指针初始化后不再替换，
// 这把锁为将来可能的替换预留；注册表内容由各实例自己的 mu 保护。
var registryMu sync.RWMutex

// ITextEncryptor 是文本加解密接口。
type ITextEncryptor interface {
	Name() string
	// 对明文做加密处理,返回密文
	Encrypt(plaintext string) (string, error)
	// 对密文做解密处理,返回明文
	Decrypt(ciphertext string) (string, error)
}

var defaultRegistry = NewEncryptorRegistry()

func init() {
	if err := defaultRegistry.Register("base64", "", Base64Encryptor); err != nil {
		panic(err)
	}
	if err := defaultRegistry.SetDefault(Base64Encryptor); err != nil {
		panic(err)
	}
}

// DefaultRegistry 返回进程级加密器注册表。
// 除 init 外，所有对全局注册表的访问都应通过本方法。
// registryMu 为将来可能的指针替换预留；当前指针初始化后不再变更。
func DefaultRegistry() *EncryptorRegistry {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return defaultRegistry
}

// GetEncrytorPrimary 返回默认加密器，等价于 DefaultRegistry().Default()。
//
// Deprecated: use DefaultRegistry().Default().
func GetEncrytorPrimary() ITextEncryptor {
	return DefaultRegistry().Default()
}

// SetEncryptorPrimary 设置默认加密器；nil 会被忽略。
//
// Deprecated: use DefaultRegistry().SetDefault().
func SetEncryptorPrimary(encryptor ITextEncryptor) {
	_ = DefaultRegistry().SetDefault(encryptor)
}
