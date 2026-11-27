package ag_crypto

import (
	"testing"

	"go.uber.org/fx"
)

func TestFxAgCryptoModuleRegisters(t *testing.T) {
	const (
		name  = "fxtest"
		keyID = "v1"
	)

	app := fx.New(
		FxAgCryptoModule,
		fx.Provide(fx.Annotate(
			func() EncryptorRegistration {
				return EncryptorRegistration{Name: name, KeyID: keyID, Enc: Base64Encryptor}
			},
			fx.ResultTags(`group:"agcrypto.encryptors"`),
		)),
	)
	if err := app.Err(); err != nil {
		t.Fatalf("fx app: %v", err)
	}
	if _, err := DefaultRegistry().Resolve(name, keyID); err != nil {
		t.Fatalf("registered encryptor not resolvable: %v", err)
	}
}
