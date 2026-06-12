package testsimple

import (
	"github.com/aif-go/ag-core/ag/ag_conf"
	"github.com/aif-go/ag-core/ag/ag_conf/reader/yaml"
	"github.com/aif-go/ag-core/ag/ag_ext"
)

var _agonet_tls_ser = `
agonet:
  server:
    addr: tcp://:8443
    config:
      security:
        type1: tls_tlcp
        type: none
        certsDir: ../certs
        TLS:
          CaPath: fgmsm/RSA_CA.cer
          AuthCertPath: fgmsm/rsa_auth_cert.cer
          AuthKeyPath: fgmsm/rsa_auth_key.pem
          SignCertPath: fgmsm/rsa_sign.cer
          SignKeyPath: fgmsm/rsa_sign_key.pem
        TLCP:
          CaPath: fgmsm/SM2_CA.cer
          AuthCertPath: fgmsm/sm2_auth_cert.cer
          AuthKeyPath: fgmsm/sm2_auth_key.pem
          SignCertPath: fgmsm/sm2_sign_cert.cer
          SignKeyPath: fgmsm/sm2_sign_key.pem
          EncCertPath: fgmsm/sm2_enc_cert.cer
          EncKeyPath: fgmsm/sm2_enc_key.pem
`

var _agonet_tls_cli = `
agonet:
  client:
    config:
      security:
        type2: tlcp
        type1: tls
        type: none
        certsDir: ../certs
        TLS:
          CaPath: fgmsm/RSA_CA.cer
          AuthCertPath: fgmsm/rsa_auth_cert.cer
          AuthKeyPath: fgmsm/rsa_auth_key.pem
          SignCertPath: fgmsm/rsa_sign.cer
          SignKeyPath: fgmsm/rsa_sign_key.pem
          InsecureSkipVerify: true
        TLCP:
          CaPath: fgmsm/SM2_CA.cer
          AuthCertPath: fgmsm/sm2_auth_cert.cer
          AuthKeyPath: fgmsm/sm2_auth_key.pem
          SignCertPath: fgmsm/sm2_sign_cert.cer
          SignKeyPath: fgmsm/sm2_sign_key.pem
          EncCertPath: fgmsm/sm2_enc_cert.cer
          EncKeyPath: fgmsm/sm2_enc_key.pem
          InsecureSkipVerify: true
`

func buildCfgBinder(confyml string) (ag_conf.IBinder, error) {
	cm, err := yaml.Read([]byte(confyml))
	if err != nil {
		return nil, err
	}
	env, _ := ag_conf.NewStandardEnvironment()
	flatmapcontext, err := ag_ext.GetFlattenedMap(cm)

	env.GetPropertySources().AddLast(&ag_conf.MapPropertySource{
		NamedPropertySource: ag_conf.NamedPropertySource{
			Name: "TEST-HZW",
		},
		Source: flatmapcontext,
	})
	binder := ag_conf.NewConfigurationPropertiesBinder(env)
	return binder, nil
}
