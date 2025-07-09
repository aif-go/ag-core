package ag_conf

import (
	"ag-core/ag/ag_crypto"
	"fmt"
	"log/slog"
	"strings"
)

// DecryptSystemConfig 遍历指定名字的properties,然后对应内容做解密处理
func DecryptSystemConfig(ps *MutablePropertySources) error {
	decryptSource := make(map[string]any, 5)
	err := ps.RangePropertySourceHandlerReverse(func(ps IPropertySource) (bool, error) { // 倒序遍历
		psname := ps.GetName()

		if strings.HasPrefix(psname, SourceKeySysPrefix) { // SYS 配置
			source := ps.GetSource()
			for key, value := range source {
				ciphertext, ok := value.(string)
				if ok && strings.HasPrefix(ciphertext, ConstEncryptKeyWords) {
					// 截取加密字符串
					ciphertext = ciphertext[len(ConstEncryptKeyWords):]
					// plaintext, err := ag_ext.GetEncrytorPrimary().Decrypt(ciphertext)
					// 此处先临时使用base64解密,后续需重构调整
					// plaintext, err := ag_crypto.Base64Encryptor.Decrypt(ciphertext)
					plaintext, err := ag_crypto.GetEncrytorPrimary().Decrypt(ciphertext)
					if err != nil {
						err = fmt.Errorf("decrypt config err source:%s, key:%s, err:%w", psname, key, err)
						slog.Error("decrypt config", "err", err)
						return true, err
					}
					decryptSource[key] = plaintext
				}
			}
		}
		return false, nil
	})
	if err != nil {
		return err
	}

	ps.AddFirst(&MapPropertySource{NamedPropertySource: NamedPropertySource{
		Name: SourceKeyDecryptSystem,
	},
		Source: decryptSource,
	})

	return nil
}

// DecryptLocalConfig 遍历local配置,做解密处理
func DecryptLocalConfig(env IConfigurableEnvironment) error {

	ps := env.GetPropertySources()

	decryptSource := make(map[string]any)
	err := ps.RangePropertySourceHandlerReverse(func(ps IPropertySource) (bool, error) {
		psname := ps.GetName()

		if strings.HasPrefix(psname, SourceKeyLocalPrefix) { // LOCAL 配置
			source := ps.GetSource()
			for key, value := range source {
				ciphertext, ok := value.(string)
				if ok && strings.HasPrefix(ciphertext, ConstEncryptKeyWords) {
					// plaintext, err := ag_ext.GetEncrytorPrimary().Decrypt(ciphertext)
					// 此处先临时使用base64解密,后续需重构调整
					// plaintext, err := ag_crypto.Base64Encryptor.Decrypt(ciphertext)
					ciphertext = ciphertext[len(ConstEncryptKeyWords):]
					plaintext, err := ag_crypto.GetEncrytorPrimary().Decrypt(ciphertext)
					if err != nil {
						err = fmt.Errorf("decrypt config err source:%s, key:%s, err:%w", psname, key, err)
						slog.Error("decrypt config", "err", err)
						return true, err
					}
					decryptSource[key] = plaintext
				}
			}
		}
		return false, nil
	})
	if err != nil {
		return err
	}

	source := &MapPropertySource{
		NamedPropertySource: NamedPropertySource{
			Name: SourceKeyDecryptLocal,
		},
		Source: decryptSource,
	}

	if ps.ContainsSource(source) {
		ps.ReplaceSource(source) // 替换
	} else if ps.Contains(SourceKeyDecryptSystem) {
		ps.AddAfter(SourceKeyDecryptSystem, source) // 添加在环境变量解密source后
	} else {
		ps.AddFirst(source) // 添加到最前面
	}
	return nil

}
