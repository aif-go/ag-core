package common

import (
	"ag-core/ag/ag_conf"
	"testing"
)

func TestNacos(t *testing.T) {
	loadEnv()
	// initNacosInfo(conf.Env)
}

func loadEnv() {
	nacosconfig := make(map[string]any, 20)
	nacosconfig["nacos.config.enable"] = "true"
	nacosconfig["nacos.config.address"] = "192.168.105.63"
	nacosconfig["nacos.config.port"] = "8813"
	nacosconfig["nacos.config.namespace"] = "aic-dev"
	nacosconfig["nacos.config.username"] = "nacos"
	nacosconfig["nacos.config.password"] = "nacos"

	// dataidlist
	nacosconfig["nacos.config.dataidlist[0].dataid"] = "aic_1.yaml"
	nacosconfig["nacos.config.dataidlist[1].dataid"] = "aic_2.yaml"
	nacosconfig["nacos.config.dataidlist[0].group"] = "DEFAULT_GROUP"
	nacosconfig["nacos.config.dataidlist[1].group"] = "DEFAULT_GROUP"
	nacosconfig["nacos.config.dataidlist[0].type"] = "yaml"
	nacosconfig["nacos.config.dataidlist[1].type"] = "yaml"

	// Timeout             uint64 `value:${"timeout_ms"} mapstructure:"nacos.config.log.timeout_ms"`
	// NotLoadCacheAtStart bool   `value:${"not_load_cache_at_start"} mapstructure:"nacos.config.log.not_load_cache_at_start"`
	nacosconfig["nacos.config.log.timeout_ms"] = "5000"
	nacosconfig["nacos.config.log.not_load_cache_at_start"] = "true"
	nacosconfig["nacos.config.log.log_dir"] = "./storage/nacos/log"
	nacosconfig["nacos.config.log.cache_dir"] = "./storage/nacos/cache"
	nacosconfig["nacos.config.log.log_level"] = "info"

	Env, _ := ag_conf.NewStandardEnvironment()
	Env.GetPropertySources().AddLast(&ag_conf.MapPropertySource{
		NamedPropertySource: ag_conf.NamedPropertySource{
			Name: "nacos",
		},
		Source: nacosconfig,
	},
	)

}

func TestParseIPPort(t *testing.T) {
	// 定义测试用例
	tests := []struct {
		name    string
		input   string
		want    []IpPort
		wantErr bool
	}{
		{
			name:  "单个IP和端口",
			input: "192.168.1.1:8080",
			want: []IpPort{
				{Ip: "192.168.1.1", Port: 8080},
			},
			wantErr: false,
		},
		{
			name:  "多个IP和端口",
			input: "192.168.1.1:8080, 10.0.0.1:8848",
			want: []IpPort{
				{Ip: "192.168.1.1", Port: 8080},
				{Ip: "10.0.0.1", Port: 8848},
			},
			wantErr: false,
		},
		{
			name:    "无效的IP地址",
			input:   "256.256.256.256:8080",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "无效的端口",
			input:   "1.2.3.4:-123",
			want:    nil,
			wantErr: true,
		},
		{
			name:  "单个IP",
			input: "192.168.1.1",
			want: []IpPort{
				{Ip: "192.168.1.1", Port: 8848},
			},
			wantErr: false,
		},
		{
			name:  "多个IP",
			input: "192.168.1.1, 10.0.0.1,3.3.3.3",
			want: []IpPort{
				{Ip: "192.168.1.1", Port: 8848},
				{Ip: "10.0.0.1", Port: 8848},
				{Ip: "3.3.3.3", Port: 8848},
			},
			wantErr: false,
		},
	}

	// 遍历测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIPPort(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIPPort() 错误 = %v, 期望错误 = %v", err, tt.wantErr)
				return
			}
			if !equal(got, tt.want) {
				t.Errorf("parseIPPort() 返回值 = %v, 期望值 = %v", got, tt.want)
			}
		})
	}
}

// 辅助函数，用于比较两个ipPort切片是否相等
func equal(a, b []IpPort) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
