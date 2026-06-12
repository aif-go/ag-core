package agsarama

import (
	"github.com/aif-go/ag-core/ag/ag_conf"

	"github.com/IBM/sarama"
	"go.uber.org/fx"
)

var FxAgsaramaModule = fx.Module("fx_agsarama_base",
	fx.Provide(
		AgsaramaModuleFx,
	),
)

// FxParams agsarama 模块的参数对象
type FxParams struct {
	fx.In

	// 配置绑定器
	Binder ag_conf.IBinder

	// 扩展配置选项
	ConfigOpts []ConfigOption `group:"agsarama",optional:"true"`
}

// FxResult agsarama 模块的结果对象
type FxResult struct {
	fx.Out

	Config       *Config        // agsarama 配置
	SaramaConfig *sarama.Config // sarama 配置
	Client       sarama.Client  // sarama 客户端
}

// AgsaramaModuleFx agsarama的 fx 模块
func AgsaramaModuleFx(params FxParams) (FxResult, error) {
	result := FxResult{}

	// 创建agsarama配置
	cfg, err := NewAgsaramaConfig(params.Binder)
	if err != nil {
		return result, err
	}
	result.Config = cfg

	// 转换agsarama配置
	sconf, err := TransConfigToSaramaConfig(cfg)
	if err != nil {
		return result, err
	}

	// 扩展agsarama配置
	copts := params.ConfigOpts
	if len(copts) > 0 {
		err := ExtendSaramaConfigWithOptions(sconf, copts...)
		if err != nil {
			return result, err
		}
	}
	result.SaramaConfig = sconf

	// 创建agsarama客户端
	client, err := NewClientWithAgConfig(cfg)
	if err != nil {
		return result, err
	}
	result.Client = client
	return result, nil
}

func FxAgsaramaGroupTag(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"agsarama"`),
	)
}
