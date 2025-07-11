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
