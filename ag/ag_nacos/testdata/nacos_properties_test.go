package test

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_conf/reader/yaml"
	"ag-core/ag/ag_ext"
	"ag-core/ag/ag_nacos/config"
	"ag-core/ag/ag_nacos/naming"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestNacosPropertiesBind(t *testing.T) {
	context, err := os.ReadFile("nacos_properties.yml")
	if err != nil {
		t.Fatal(err)
	}

	contextMap, err := yaml.Read(context)
	if err != nil {
		t.Fatal(err)
	}
	flatmapcontext, err := ag_ext.GetFlattenedMap(contextMap)
	if err != nil {
		t.Fatal(err)
	}
	env, _ := ag_conf.NewStandardEnvironment()
	// 如果localProperties已经存在,此时需要将yaml添加到它之前
	env.GetPropertySources().AddLast(&ag_conf.MapPropertySource{
		NamedPropertySource: ag_conf.NamedPropertySource{
			Name: "nacos_properties_test",
		},
		Source: flatmapcontext,
	})

	binder := ag_conf.NewConfigurationPropertiesBinder(env)

	np, _ := naming.NewNacosNamingProperties(binder)
	jstr, _ := json.MarshalIndent(np, "", " ")
	fmt.Printf("np:%s\n", jstr)

	cp, _ := config.NewNacosConfigProperties(binder)
	jstr, _ = json.MarshalIndent(cp, "", " ")
	fmt.Printf("cp:%s\n", jstr)
}
