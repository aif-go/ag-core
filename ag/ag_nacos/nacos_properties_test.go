package ag_nacos

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_conf/reader/yaml"
	"ag-core/ag/ag_ext"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestNacosPropertiesBind(t *testing.T) {
	context, err := os.ReadFile("./testdata/nacos_properties.yml")
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

	nacosProperties := &NacosConfigProperties{}

	err = binder.Bind(nacosProperties, NacosConfigPropertiesPrefix)
	if err != nil {
		t.Fatal(err)
	}
	jstr, err := json.MarshalIndent(nacosProperties, "", " ")
	fmt.Printf("err:%v %s", err, jstr)

}
