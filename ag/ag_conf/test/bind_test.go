package test

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_conf/reader/yaml"
	"ag-core/ag/ag_ext"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestConfigBinder(t *testing.T) {
	_, binder := getEnvBinder(t)
	hzwArrayss := &HzwArrayss{}

	err := binder.Bind(hzwArrayss, "hzwarrayss")
	if err != nil {
		t.Fatalf("bind config failed, err: %v", err)
	}
	showJson(t, hzwArrayss)

}

func getEnvBinder(t *testing.T) (*ag_conf.StandardEnvironment, *ag_conf.ConfigurationPropertiesBinder) {
	yamlmap := getyamlmap(t)
	return getEnvAndBinderByMap(t, yamlmap)
}

func getyamlmap(t *testing.T) map[string]any {
	bytearr, err := os.ReadFile("app.yml")
	if err != nil {
		t.Fatalf("read yaml file failed, err: %v", err)
	}
	yamlmap, err := yaml.Read(bytearr)
	if err != nil {
		t.Fatalf("read yaml file failed, err: %v", err)
	}
	return yamlmap
}

func getEnvAndBinderByMap(t *testing.T, yamlmap map[string]any) (*ag_conf.StandardEnvironment, *ag_conf.ConfigurationPropertiesBinder) {
	env, _ := ag_conf.NewStandardEnvironment()
	flatmapcontext, err := ag_ext.GetFlattenedMap(yamlmap)
	if err != nil {
		t.Fatalf("get flattened map failed, err: %v", err)
	}

	env.GetPropertySources().AddLast(&ag_conf.MapPropertySource{
		NamedPropertySource: ag_conf.NamedPropertySource{
			Name: "TEST-HZW",
		},
		Source: flatmapcontext,
	})
	binder := ag_conf.NewConfigurationPropertiesBinder(env)

	return env, binder
}

func showJson(t *testing.T, v any) {
	jsonstr, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("to json string failed, err: %v", err)
	}
	fmt.Printf("json: %T %s", v, jsonstr)
}
