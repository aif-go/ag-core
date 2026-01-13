package test

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_conf/reader/yaml"
	"ag-core/ag/ag_ext"
	"ag-core/contribute/agredis"
	"encoding/json"
	"fmt"
	"testing"
)

var (
	_single string = `
agredis:
  config:
    addrs:
      - "127.0.0.1:6379"
`
	_rw string = `
agredis:
  type: rw
  config:
    addrs:
      - "127.0.0.1:6379"
  Replicas:
    - addrs:
        - "127.0.0.1:6379"
`
)

func TestBindAgRedisProperties(t *testing.T) {
	// confyml := _single
	confyml := _rw
	cm, err := yaml.Read([]byte(confyml))
	if err != nil {
		t.Fatal(err)
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

	conf := &agredis.AgRedisProperties{}
	// conf := &AgRedisProperties{
	// 	Config:  &AgUniversalOptionsProperties{},
	// 	Replica: &AgUniversalOptionsProperties{},
	// }
	err = binder.Bind(conf, agredis.AgRedisConfPrefix)
	if err != nil {
		t.Fatal(err)
	}

	confjson, _ := json.MarshalIndent(conf, "", " ")
	fmt.Println(string(confjson))

	build := agredis.AgRedisClientBuilder{
		Config: conf,
	}

	cli, err := build.Build()
	if err != nil {
		t.Fatal(err)
	}

	testCase1(cli, t)

}
