package agredis

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_conf/reader/yaml"
	"ag-core/ag/ag_ext"
	"encoding/json"
	"fmt"
	"testing"
)

var content string = `
agredis:
  type: rw
  config:
    db: 2
    ClientName: xxxx
    addrs:
      - "127.0.0.1:6379"
    MinRetryBackoff: 5
  Replicas:
    - db: 2
      ClientName: xxxx
      addrs:
        - "127.0.0.1:6380"
      MinRetryBackoff: 5
  testarray:
    - - 1
      - 2
`

func TestBindAgRedisProperties(t *testing.T) {
	cm, err := yaml.Read([]byte(content))
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

	conf := &AgRedisProperties{}
	// conf := &AgRedisProperties{
	// 	Config:  &AgUniversalOptionsProperties{},
	// 	Replica: &AgUniversalOptionsProperties{},
	// }
	err = binder.Bind(conf, AgRedisConfPrefix)
	if err != nil {
		fmt.Println("=====")
		t.Fatal(err)
	}

	confjson, _ := json.MarshalIndent(conf, "", " ")
	fmt.Println(string(confjson))

}
