package ag_conf_test

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_conf/reader/yaml"
	"ag-core/ag/ag_ext"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"testing"
	"time"
)

type Hzw struct {
	Name string `value:"${name:111}"`
	Age  int    `value:"${age:22}"`
}

type Extra struct {
	Bool     bool           `value:"${bool:true}" `
	Int      int            `value:"${int:4}" `
	Int8     int8           `value:"${int8:8}" `
	Int16    int16          `value:"${int16:16}" `
	Int32    int32          `value:"${int32:32}" `
	Int64    int64          `value:"${int32:64}" `
	Uint     uint           `value:"${uint:4}" `
	Uint8    uint8          `value:"${uint8:8}" `
	Uint16   uint16         `value:"${uint16:16}" `
	Uint32   uint32         `value:"${uint32:32}" `
	Uint64   uint64         `value:"${uint32:64}" `
	Float32  float32        `value:"${float32:3.2}" `
	Float64  float64        `value:"${float64:6.4}" `
	String   string         `value:"${string:xyz}" `
	string2  string         `value:"${string2:hhhh}"`
	Duration time.Duration  `value:"${duration:10}"`
	IntsV0   []int          `value:"${intsV0:}"`
	IntsV1   []int          `value:"${intsV1:1,2,3}"`
	IntsV2   []int          `value:"${intsV2}"`
	MapV0    map[string]int `value:"${mapV0:}"`
	MapV2    map[string]int `value:"${mapV2}"`
	Hzw                     // Anonymous field 匿名字段
}

func TestConfBind(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	// var int1 int
	// Bind(Env, &int1, "int")

	hzwps := &ag_conf.MapPropertySource{}
	hzwps.Name = "hzw"
	hzwps.Source = map[string]any{
		"string":         "hzwhome",
		"int":            "123",
		"uint":           "123",
		"bool":           "true",
		"boolerr":        "true1",
		"dur":            "1000",
		"arr[0]":         "11",
		"arr[1]":         "22",
		"arr[2]":         "33",
		"arr2[0].string": "11",
		"arr2[1].string": "22",
		"arr2[2].string": "33",
		"mapV0.a":        "1",
		"mapV2.a":        "2",
	}
	env, _ := ag_conf.NewStandardEnvironment()
	env.GetPropertySources().AddFirst(hzwps)

	binder := ag_conf.NewConfigurationPropertiesBinder(env)

	var str string
	binder.Bind(&str, "string")
	fmt.Printf("str=%s\n", str)
	var int1 int
	binder.Bind(&int1, "int")
	fmt.Printf("int1=%d\n", int1)
	var uint1 uint
	binder.Bind(&uint1, "uint")
	fmt.Printf("uint1=%d\n", uint1)
	var bool1 bool
	binder.Bind(&bool1, "bool")
	fmt.Printf("bool1=%v\n", bool1)

	err := binder.Bind(&bool1, "boolerr")
	fmt.Printf("err=%v\n", err)

	var duration time.Duration
	err = binder.Bind(&duration, "dur")
	fmt.Printf("err: %v Duration=%v\n", err, duration)

	var ext Extra
	err = binder.Bind(&ext)
	if err != nil {
		t.Errorf("err: %v\n", err)
	}
	extjson, _ := json.MarshalIndent(&ext, " ", " ")
	fmt.Printf("ext=%s\n", extjson)

	// TODO bind nil测试

	var slice1 []int
	binder.Bind(&slice1, "arr")
	fmt.Printf("slice1=%v\n", slice1)

}

func TestConfBindSlice(t *testing.T) {
	hzwps := &ag_conf.MapPropertySource{}
	hzwps.Name = "hzw"
	hzwps.Source = map[string]any{
		"hzw[0].name":   "11",
		"hzw[1].name":   "22",
		"hzw[2].name":   "33",
		"ext[0].string": "11",
		"ext[1].string": "22",
		"ext[2].string": "33",
	}
	env, _ := ag_conf.NewStandardEnvironment()
	env.GetPropertySources().AddFirst(hzwps)

	binder := ag_conf.NewConfigurationPropertiesBinder(env)

	var slice1 []Hzw
	err := binder.Bind(&slice1, "hzw")
	if err != nil {
		t.Errorf("err: %v\n", err)
	}
	fmt.Printf("slice1=%v\n", slice1)

	slice2 := make([]Hzw, 0)
	binder.Bind(&slice2, "hzw")
	fmt.Printf("slice2=%v\n", slice2)

	slice3 := make([]*Hzw, 0)
	fmt.Printf("slice3 T %T\n", slice3)
	err = binder.Bind(&slice3, "hzw")
	fmt.Printf("err:%v, slice3=%v\n", err, slice3) // 空

	var slice4 []Extra
	binder.Bind(&slice4, "ext")
	fmt.Printf("slice4=%v\n", slice4)

}

func TestConfBindMap(t *testing.T) {
	hzwps := &ag_conf.MapPropertySource{}
	hzwps.Name = "hzw"
	hzwps.Source = map[string]any{
		"hzwmap.hzw1.name": "11",
		"hzwmap.hzw1.age":  "11",
		"hzwmap.hzw2.name": "22",
		"hzwmap.hzw2.age":  "22",
		"map2.aa":          "AA",
		"map2.bb":          "BB",
	}
	env, _ := ag_conf.NewStandardEnvironment()
	env.GetPropertySources().AddFirst(hzwps)

	binder := ag_conf.NewConfigurationPropertiesBinder(env)

	hzwmap := make(map[string]Hzw)
	err := binder.Bind(&hzwmap, "hzwmap")
	fmt.Printf("err:%v hzwmap=%v\n", err, hzwmap)

	map2 := make(map[string]string)
	binder.Bind(&map2, "map2")
	fmt.Printf("err:%v map2=%v\n", err, map2)

}

func TestConfBindMapSlice(t *testing.T) {
	content := `
HZW:
  KMAP:
    k1:
      -
        - "k1s1-1"
        - "k1s1-2"
      -
        - "k1s2-1"
        - "k1s2-2"
    k2:
      -
        - "k2s1-1"
      - 
  
`

	type K struct {
		KMAP map[string][][]string
	}
	// type K2 struct {
	// 	KMAP map[string]map[string][]string
	// }

	cm, err := yaml.Read([]byte(content))
	if err != nil {
		t.Fatal(err)
	}

	flatmapcontext, err := ag_ext.GetFlattenedMap(cm)
	fjson, _ := json.MarshalIndent(flatmapcontext, " ", " ")
	fmt.Printf("fjson=%s\n\n", fjson)

	if err != nil {
		t.Fatal(err)
	}
	env, _ := ag_conf.NewStandardEnvironment()

	env.GetPropertySources().AddLast(&ag_conf.MapPropertySource{
		NamedPropertySource: ag_conf.NamedPropertySource{
			Name: "TEST-HZW",
		},
		Source: flatmapcontext,
	})
	binder := ag_conf.NewConfigurationPropertiesBinder(env)

	var k K
	binder.Bind(&k, "HZW")
	jsonk, _ := json.MarshalIndent(&k, " ", " ")
	fmt.Printf("k=%s\n\n", jsonk)

	// var k2 K2
	// binder.Bind(&k2, "HZW")
	// jsonk2, _ := json.MarshalIndent(&k2, " ", " ")
	// fmt.Printf("k2=%s\n\n", jsonk2)
	// k2k1_0, ok := k2.KMAP["k1"]["[0]"]
	// if !ok {
	// 	t.Error("k2k1l not found")
	// }
	// if len(k2k1_0) != 2 {
	// 	t.Error("k2k1_0 len not 2")
	// }

}

func TestParseTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		want    ag_conf.ParsedTag
		wantErr bool
	}{
		{
			name: "normal with default",
			tag:  "${key:value}",
			want: ag_conf.ParsedTag{
				Key:    "key",
				Def:    "value",
				HasDef: true,
				// Splitter: "split",
			},
			wantErr: false,
		},
		{
			name: "normal without default",
			tag:  "${key}",
			want: ag_conf.ParsedTag{
				Key:    "key",
				Def:    "",
				HasDef: false,
				// Splitter: "split",
			},
			wantErr: false,
		},
		{
			name: "no splitter",
			tag:  "${key:value}",
			want: ag_conf.ParsedTag{
				Key:    "key",
				Def:    "value",
				HasDef: true,
				// Splitter: "",
			},
			wantErr: false,
		},
		{
			name:    "empty tag",
			tag:     "",
			want:    ag_conf.ParsedTag{},
			wantErr: true,
		},
		{
			name:    "invalid format",
			tag:     "${key",
			want:    ag_conf.ParsedTag{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ag_conf.ParseTag(tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanSet(t *testing.T) {
	map1 := make(map[string]int)
	fmt.Printf("%T canset=%v kind:%v\n", map1, reflect.ValueOf(map1).CanSet(), reflect.ValueOf(map1).Kind())
	map2 := &map1
	fmt.Printf("%T ecanset=%v kind:%v\n", map2, reflect.ValueOf(map2).Elem().CanSet(), reflect.ValueOf(map2).Kind())

	slice1 := make([]int, 0)
	fmt.Printf("%T canset=%v kind:%v\n", slice1, reflect.ValueOf(slice1).CanSet(), reflect.ValueOf(slice1).Kind())
	slice2 := &slice1
	fmt.Printf("%T ecanset=%v kind:%v\n", slice2, reflect.ValueOf(slice2).Elem().CanSet(), reflect.ValueOf(slice2).Kind())

	array1 := [3]int{}
	fmt.Printf("%T canset=%v kind:%v\n", array1, reflect.ValueOf(array1).CanSet(), reflect.ValueOf(array1).Kind())
	array2 := &array1
	fmt.Printf("%T ecanset=%v kind:%v\n", array2, reflect.ValueOf(array2).Elem().CanSet(), reflect.ValueOf(array2).Kind())

	i1 := 1
	fmt.Printf("%T canset=%v kind:%v\n", i1, reflect.ValueOf(i1).CanSet(), reflect.ValueOf(i1).Kind())
	i2 := &i1
	fmt.Printf("%T ecanset=%v kind:%v\n", i2, reflect.ValueOf(i2).Elem().CanSet(), reflect.ValueOf(i2).Kind())

	mps1 := ag_conf.MapPropertySource{}
	fmt.Printf("%T canset=%v kind:%v\n", mps1, reflect.ValueOf(mps1).CanSet(), reflect.ValueOf(mps1).Kind())
	mps2 := &mps1
	fmt.Printf("%T ecanset=%v kind:%v\n", mps2, reflect.ValueOf(mps2).Elem().CanSet(), reflect.ValueOf(mps2).Kind())
}
