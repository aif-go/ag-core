package ag_conf

import (
	"github.com/aif-go/ag-core/ag/ag_crypto"
	"fmt"
	"log/slog"
	"strings"
)

// DecryptLocalConfig 遍历local配置,做解密处理
func DecryptLocalConfig(env IConfigurableEnvironment) error {
	pss := env.GetPropertySources()

	err := pss.RangePropertySourceHandlerReverse(func(ps IPropertySource) (bool, error) {
		psname := ps.GetName()
		if strings.HasPrefix(psname, SourceKeyLocalPrefix) { // LOCAL 配置
			err := CreateOrUpdateDecryptForPropertySource(env, ps)
			if err != nil {
				return true, err
			}
		}
		return false, nil
	})
	if err != nil {
		return err
	}

	return nil

}

func CreateOrUpdateDecryptForPropertySource(env IConfigurableEnvironment, ps IPropertySource) error {
	pss := env.GetPropertySources()
	if pss == nil {
		return fmt.Errorf("env getPropertySources is nil")
	}

	psname := ps.GetName()
	dsource, err := BuildDecryptForPropertySource(ps)
	if err != nil {
		slog.Error("decrypt config", "err", err)
		return err
	}

	if len(dsource.Source) > 0 {
		err = pss.AddBefore(psname, dsource)
		if err != nil {
			slog.Error("decrypt config", "err", err)
			return err
		}
	} else {
		pss.RemoveIfPresent(dsource.GetName())
	}

	// if len(dsource.Source) <= 0 {
	// 	pss.RemoveIfPresent(dsource.GetName())
	// } else if pss.ContainsSource(dsource) {
	// 	pss.ReplaceSource(dsource) // 替换
	// } else {
	// 	// pss.AddFirst(dsource) // 添加到最前面
	// 	pss.AddBefore(psname, dsource) // 添加到目标LocalSource之前
	// }
	return nil
}

func BuildDecryptForPropertySource(ps IPropertySource) (*MapPropertySource, error) {
	psname := ps.GetName()
	if strings.HasPrefix(psname, SourceKeyDecryptPrefix) {
		return nil, fmt.Errorf("decrypt config err source:%s, name:%s is already decrypt", psname, psname)
	}

	decryptSource := make(map[string]any)
	source := ps.GetSource()

	encryptor := ag_crypto.GetEncrytorPrimary()
	// encname := encryptor.Name()

	for key, value := range source {
		ciphertext, ok := value.(string)
		if ok && strings.HasPrefix(ciphertext, ConstEncryptKeyWords) {
			// plaintext, err := ag_ext.GetEncrytorPrimary().Decrypt(ciphertext)
			// 此处先临时使用base64解密,后续需重构调整
			// plaintext, err := ag_crypto.Base64Encryptor.Decrypt(ciphertext)
			ciphertext = ciphertext[len(ConstEncryptKeyWords):]
			plaintext, err := encryptor.Decrypt(ciphertext)
			if err != nil {
				err = fmt.Errorf("decrypt config err source:%s, key:%s, err:%w", psname, key, err)
				return nil, err
			}
			decryptSource[key] = plaintext
		}
	}

	dsource := &MapPropertySource{
		NamedPropertySource: NamedPropertySource{
			// Name: fmt.Sprintf("%s_[%s]_%s", SourceKeyDecryptPrefix, encname, psname),
			Name: fmt.Sprintf("%s_%s", SourceKeyDecryptPrefix, psname),
		},
		Source: decryptSource,
	}

	return dsource, nil
}
