package ag_crypto

var DefaultEncryptorPrimary ITextEncryptor = Base64Encryptor

type ITextEncryptor interface {
	Name() string
	// 对明文做加密处理,返回密文
	Encrypt(plaintext string) (string, error)
	// 对密文做解密处理,返回明文
	Decrypt(ciphertext string) (string, error)
}

func GetEncrytorPrimary() ITextEncryptor {
	return DefaultEncryptorPrimary
}

func SetEncryptorPrimary(encryptor ITextEncryptor) {
	DefaultEncryptorPrimary = encryptor
}
