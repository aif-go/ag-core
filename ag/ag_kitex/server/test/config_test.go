package kitex_test

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_ext"
	kitex "ag-core/ag/ag_kitex/server"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConfBind(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	path := "./config.yml"
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// 将yaml解析为map
	var rs map[string]any
	yaml.Unmarshal(bytes, &rs)

	// 将map构建为扁平的map
	rr, err := ag_ext.GetFlattenedMap(rs)
	if err != nil {
		t.Fatal(err)
	}

	source := &ag_conf.MapPropertySource{}
	source.Name = "hzw"
	source.Source = rr

	env, _ := ag_conf.NewStandardEnvironment()
	env.GetPropertySources().AddFirst(source)

	binder := ag_conf.NewConfigurationPropertiesBinder(env)

	var p kitex.KitexServerProperties
	pv := reflect.ValueOf(&p)
	pv = pv.Elem() // 获取指针指向的元素
	binder.Bind(pv, kitex.KitexServerPropertiesPrefix)
	pjson, err := json.MarshalIndent(p, " ", "  ")
	fmt.Printf("%s\n", pjson)

	var p2 kitex.KitexServerProperties
	binder.Bind(&p2, kitex.KitexServerPropertiesPrefix)
	pjson, err = json.MarshalIndent(p2, " ", "  ")
	fmt.Printf("%s\n", pjson)

	nilv := reflect.ValueOf(nil)
	err = binder.Bind(nilv, kitex.KitexServerPropertiesPrefix)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	var p3 *kitex.KitexServerProperties
	err = binder.Bind(p3, kitex.KitexServerPropertiesPrefix)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	p4 := &kitex.KitexServerProperties{}
	err = binder.Bind(p4, kitex.KitexServerPropertiesPrefix)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	pjson, err = json.MarshalIndent(p4, " ", "  ")
	fmt.Printf("%s\n", pjson)
}
