package config

import (
	"ag-core/ag/ag_conf"
	"ag-core/contribute/agnacos/common"
	"errors"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

func NewNacosConfigProperties(binder ag_conf.IBinder) (*NacosConfigProperties, error) {
	p := &NacosConfigProperties{}
	err := binder.Bind(p, NacosConfigPropertiesPrefix)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func NewNacosConfigClient(p *NacosConfigProperties) (config_client.IConfigClient, error) {

	if p == nil || !p.Enable {
		return nil, nil
	}

	sc, err := common.BuildServerConfig(p.SCProperties)
	if err != nil {
		return nil, err
	}
	if len(sc) == 0 {
		return nil, errors.New("nacos server config is empty")
	}

	// cc, err := namingClientConfig(p)
	cc, err := common.BuildClientConfig(p.SCProperties)
	if err != nil {
		return nil, err
	}
	if cc == nil {
		return nil, errors.New("nacos client config is empty")
	}

	cli, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  cc,
			ServerConfigs: sc,
		},
	)

	if err != nil {
		return nil, err
	}
	return cli, nil
}
