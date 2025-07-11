package test

import (
	"ag-core/ag/ag_conf"
	"fmt"
	"log/slog"
	"testing"
	"time"
)

func TestEnvironment(t *testing.T) {

	slog.SetLogLoggerLevel(slog.LevelDebug)
	// var int1 int
	// Bind(Env, &int1, "int")

	env, _ := ag_conf.NewStandardEnvironment()

	hzwps := &ag_conf.MapPropertySource{}
	hzwps.Name = "hzw"
	hzwps.Source = map[string]any{
		"hhh": "HHHHHHHH",
		"zzz": "${hhh2}",
		"www": "${zzz}",

		"h":   "H",
		"z":   "Z",
		"w":   "W",
		"hzw": "${h}_${z}_${w}_${y:yy}",

		"xxx1": "${xxxxxxx}",

		"xxx1_2": "${h${z${w}}}", // = ${h${zW}} =  ${hZW}
		"zW":     "ZW",
		"hZW":    "HZW",

		"c1": "${c2}",
		"c2": "${c3}",
		"c3": "${c1:111}",
	}
	env.GetPropertySources().AddFirst(hzwps)

	hzwps2 := &ag_conf.MapPropertySource{}
	hzwps2.Name = "hzw2"
	hzwps2.Source = map[string]any{
		"hhh2": "HHHHHHHH2",
		"zzz2": "${hhh2}",
		"www2": "${zzz2}",
		"xxx2": "${xxxxxxx}",
	}
	env.GetPropertySources().AddFirst(hzwps2)

	v := env.GetProperty("xxx1")
	fmt.Println(v)

	time.Sleep(time.Millisecond)
	fmt.Println("================")
	v = env.GetProperty("xxx1_2")
	fmt.Println(v)

	time.Sleep(time.Millisecond)
	fmt.Println("================")
	v = env.GetProperty("xxx2")
	fmt.Println(v)

	time.Sleep(time.Millisecond)
	fmt.Println("========c1========")
	v = env.GetProperty("c1")
	fmt.Println(v)

	time.Sleep(time.Millisecond)
	fmt.Println("================")
	v = env.GetProperty("hzw")
	fmt.Println(v)

}

func TestPropertySource(t *testing.T) {
	t.Run("测试属性源初始添加", func(t *testing.T) {
		env, _ := ag_conf.NewStandardEnvironment()
		pss := env.GetPropertySources()

		// 初始添加属性源
		hzwps := &ag_conf.MapPropertySource{}
		hzwps.Name = "hzw"
		hzwps.Source = map[string]any{
			"h": "1",
		}
		env.GetPropertySources().AddFirst(hzwps)
		sourceSlice := pss.GetPropertySources()

		// 验证初始值
		s1 := sourceSlice[0]
		vh := s1.GetProperty("h")
		fmt.Printf("初始h值:%v \n", vh)
		if vh != "1" {
			t.Errorf("期望初始h值为1，实际得到:%v", vh)
		}
	})

	t.Run("测试属性源替换", func(t *testing.T) {
		env, _ := ag_conf.NewStandardEnvironment()
		pss := env.GetPropertySources()

		// 初始属性源
		hzwps := &ag_conf.MapPropertySource{}
		hzwps.Name = "hzw"
		hzwps.Source = map[string]any{
			"h": "1",
		}
		env.GetPropertySources().AddFirst(hzwps)
		sourceSlice := pss.GetPropertySources()

		// 替换属性源
		hzwps2 := &ag_conf.MapPropertySource{}
		hzwps2.Name = "hzw"
		hzwps2.Source = map[string]any{
			"h": "2",
		}
		env.GetPropertySources().ReplaceSource(hzwps2)

		// 验证旧切片不变
		s1 := sourceSlice[0]
		vh := s1.GetProperty("h")
		fmt.Printf("替换后旧切片h值:%v \n", vh)
		if vh != "1" {
			t.Errorf("期望旧切片h值保持1，实际得到:%v", vh)
		}

		// 验证新切片已更新
		s1 = pss.GetPropertySources()[0]
		vh = s1.GetProperty("h")
		fmt.Printf("替换后新切片h值:%v \n", vh)
		if vh != "2" {
			t.Errorf("期望新切片h值更新为2，实际得到:%v", vh)
		}
	})

	t.Run("测试属性源直接修改", func(t *testing.T) {
		env, _ := ag_conf.NewStandardEnvironment()
		pss := env.GetPropertySources()

		// 初始属性源
		hzwps := &ag_conf.MapPropertySource{}
		hzwps.Name = "hzw"
		hzwps.Source = map[string]any{
			"h": "1",
		}
		env.GetPropertySources().AddFirst(hzwps)
		sourceSlice := pss.GetPropertySources()

		// 直接修改源数据
		hzwps.Source["h"] = "11"

		// 验证修改后的值
		s1 := sourceSlice[0]
		vh := s1.GetProperty("h")
		fmt.Printf("直接修改后h值:%v \n", vh)
		if vh != "11" {
			t.Errorf("期望直接修改后h值为11，实际得到:%v", vh)
		}
	})
}
