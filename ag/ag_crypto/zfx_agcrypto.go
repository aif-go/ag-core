package ag_crypto

import "go.uber.org/fx"

// EncryptorRegistration 是插件向 FX 注册加密器的描述。
// 插件用 fx.Annotate(..., fx.ResultTags(`group:"agcrypto.encryptors"`)) 提供它。
type EncryptorRegistration struct {
	Name  string
	KeyID string
	Enc   ITextEncryptor
}

type cryptoRegistrations struct {
	fx.In
	Items []EncryptorRegistration `group:"agcrypto.encryptors"`
}

// FxAgCryptoModule 把插件注册的加密器聚合到进程级注册表。
// 它提供的是 DefaultRegistry() 同一实例，ag_conf 无需额外接线。
var FxAgCryptoModule = fx.Module(
	"agcrypto",
	fx.Provide(func() *EncryptorRegistry { return DefaultRegistry() }),
	fx.Invoke(func(reg *EncryptorRegistry, in cryptoRegistrations) error {
		for _, item := range in.Items {
			if err := reg.Register(item.Name, item.KeyID, item.Enc); err != nil {
				return err
			}
		}
		return nil
	}),
)
