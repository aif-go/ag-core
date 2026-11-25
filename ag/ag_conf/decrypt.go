package ag_conf

import (
	"fmt"
	"github.com/aif-go/ag-core/ag/ag_crypto"
	"github.com/aif-go/ag-core/ag/ag_crypto/protocol"
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

	for key, value := range source {
		ciphertext, ok := value.(string)
		if !ok {
			continue
		}

		ref, isCipher, err := protocol.Parse(ciphertext)
		if err != nil {
			return nil, fmt.Errorf("decrypt config err source:%s, key:%s: %w", psname, key, err)
		}
		if !isCipher {
			continue
		}
		var enc ag_crypto.ITextEncryptor
		if ref.Name == "" {
			// 兼容旧 API：默认加密器仍以 DefaultEncryptorPrimary 为准
			enc = ag_crypto.GetEncrytorPrimary()
			if enc == nil {
				return nil, fmt.Errorf("decrypt config err source:%s, key:%s: no default encryptor", psname, key)
			}
		} else {
			resolved, rerr := ag_crypto.DefaultRegistry().Resolve(ref.Name, ref.KeyID)
			if rerr != nil {
				return nil, fmt.Errorf("decrypt config err source:%s, key:%s, encryptor:%s/%s: %w",
					psname, key, ref.Name, ref.KeyID, rerr)
			}
			enc = resolved
		}
		plaintext, err := enc.Decrypt(ref.Payload)
		if err != nil {
			return nil, fmt.Errorf("decrypt config err source:%s, key:%s, encryptor:%s/%s: %w",
				psname, key, ref.Name, ref.KeyID, err)
		}
		decryptSource[key] = plaintext
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
