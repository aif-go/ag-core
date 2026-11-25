package ag_crypto

// DefaultEncryptorPrimary 是历史上的全局单例。
//
// Deprecated: 直接赋值不会同步到注册表，请改用 SetEncryptorPrimary 或 DefaultRegistry().SetDefault()。
var DefaultEncryptorPrimary ITextEncryptor = Base64Encryptor

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
func DefaultRegistry() *EncryptorRegistry {
	return defaultRegistry
}

// GetEncrytorPrimary 返回默认加密器。
//
// Deprecated: use DefaultRegistry().Default().
func GetEncrytorPrimary() ITextEncryptor {
	return DefaultEncryptorPrimary
}

// SetEncryptorPrimary 设置默认加密器。
//
// Deprecated: use DefaultRegistry().SetDefault().
func SetEncryptorPrimary(encryptor ITextEncryptor) {
	DefaultEncryptorPrimary = encryptor
	_ = defaultRegistry.SetDefault(encryptor)
}
