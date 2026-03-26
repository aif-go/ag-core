package testtls

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_conf/reader/yaml"
	"ag-core/ag/ag_ext"
	"ag-core/contribute/agonet"
	"testing"
)

var _agonet_tls_ser = `
agonet:
  server:
    addr: tcp://:8443
    config:
      security:
        type: tls_tlcp
        type1: none
        certsDir: ./certs
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

func TestAgCfgTlsTlcp_Server(t *testing.T) {
	binder, err := buildCfgBinder(_agonet_tls_ser)
	if err != nil {
		t.Fatalf("buildCfgBinder failed: %v", err)
	}

	serverCfg, err := agonet.NewServerConfig(binder)
	if err != nil {
		t.Fatalf("NewServerConfig failed: %v", err)
	}

	handler := &TestEventHandler{}

	server, err := agonet.NewServer(handler, serverCfg)
	// server, err := agonet.NewServerWithOptions(handler, []string{"tcp://:8443"}, opts)
	if err != nil {
		t.Fatalf("NewServerWithOptions failed: %v", err)
	}

	err = server.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
}

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
