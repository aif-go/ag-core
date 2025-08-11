package test

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_ext"
	"ag-core/ag/ag_log/slogzap"
	"fmt"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSlogZapConf(t *testing.T) {

	bytearr, err := os.ReadFile("test.yaml")
	if err != nil {
		panic(err)
	}
	mapcontext := make(map[string]any)
	err = yaml.Unmarshal(bytearr, mapcontext)
	if err != nil {
		panic(err)
	}

	mapcontext, err = ag_ext.GetFlattenedMap(mapcontext)
	if err != nil {
		panic(err)
	}

	env, _ := ag_conf.NewStandardEnvironment()
	mpp := &ag_conf.MapPropertySource{
		NamedPropertySource: ag_conf.NamedPropertySource{
			Name: "logtest",
		},
		Source: mapcontext,
	}
	env.GetPropertySources().AddLast(mpp)

	binder := ag_conf.NewConfigurationPropertiesBinder(env)

	opt := &slogzap.SlogZapProperties{}
	err = binder.Bind(opt, "aglog.zap")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(opt)

}
