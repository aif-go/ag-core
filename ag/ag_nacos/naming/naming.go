package naming

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_nacos/common"
	"errors"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

func NewNacosNamingProperties(binder ag_conf.IBinder) (*NacosNamingProperties, error) {
	p := &NacosNamingProperties{}
	err := binder.Bind(p, NacosNamingPropertiesPrefix)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func NewNacosNamingClient(p *NacosNamingProperties) (naming_client.INamingClient, error) {

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

	cc, err := common.BuildClientConfig(p.SCProperties)
	if err != nil {
		return nil, err
	}
	if cc == nil {
		return nil, errors.New("nacos client config is empty")
	}

	cli, err := clients.NewNamingClient(
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
