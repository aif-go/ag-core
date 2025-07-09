package ag_nacos

import (
	// "fmt"
	// "log/slog"

	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_conf/reader/json"
	"ag-core/ag/ag_conf/reader/prop"
	"ag-core/ag/ag_conf/reader/yaml"
	"ag-core/ag/ag_ext"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"

	// "wiredemo/pkg/conf"
	// "wiredemo/pkg/conf/reader/json"
	// "wiredemo/pkg/conf/reader/prop"
	// "wiredemo/pkg/conf/reader/yaml"

	// "wiredemo/pkg/conf/reader/json"
	// "wiredemo/pkg/conf/reader/prop"
	// "wiredemo/pkg/conf/reader/yaml"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/spf13/cast"
)

// NacosInfo nacos config
// var NacosInfo *model.NacosInfo = &model.NacosInfo{}

// // NacosDataIDList dataid collections
// var NacosDataIDList *model.NacosDataIDList = &model.NacosDataIDList{}

// // NacosLog  nacos log config
// var NacosLog *model.NacosLog = &model.NacosLog{}

// type AutoconfigNacosConfig string

// // NacosBaseInfo nacos base info
// type NacosBaseInfo struct {

// 	// NacosClientConfig nacos client config
// 	NacosClientConfig *constant.ClientConfig

// 	// NacosServerConfigSlice  nacos server config
// 	NacosServerConfigSlice []constant.ServerConfig
// 	// NacosNamingClient nacos naming client
// 	NacosNamingClient naming_client.INamingClient

// 	// NacosConfigClient nacos config client
// 	NacosConfigClient config_client.IConfigClient
// }

// NacosBaseInfo
// var baseInfo *NacosBaseInfo

// func initNacosInfo(env conf.IConfigurableEnvironment) {
// 	// 如果未开启nacos,后续项都不需要加载,默认关闭
// 	if !EnableNacosConfig() {
// 		return
// 	}
// 	// 先获取本地资源信息
// 	err := conf.Binder.Bind(NacosInfo, "nacos.config")
// 	if err != nil {
// 		panic(err)
// 	}
// 	err = conf.Binder.Bind(NacosLog, "nacos.config.log")
// 	if err != nil {
// 		panic(err)
// 	}
// 	err = conf.Binder.Bind(NacosDataIDList, "nacos.config")
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func NewNacosInfo(bind conf.IBinder) (*model.NacosInfo, error) {
// 	nacosInfo := &model.NacosInfo{}
// 	err := bind.Bind(nacosInfo, "nacos.config")
// 	return nacosInfo, err
// }
// func NewNacosLog(bind conf.IBinder) (*model.NacosLog, error) {
// 	nacosLog := &model.NacosLog{}
// 	err := bind.Bind(nacosLog, "nacos.config.log")
// 	return nacosLog, err
// }
// func NewNacosDataIDList(bind conf.IBinder) (*model.NacosDataIDList, error) {
// 	nacosDataIDList := &model.NacosDataIDList{}
// 	err := bind.Bind(nacosDataIDList, "nacos.config")
// 	return nacosDataIDList, err
// }

// NewNacosProperties 构建nacos配置参数
func NewNacosProperties(binder ag_conf.IBinder) (*NacosConfigProperties, error) {
	nacosProperties := &NacosConfigProperties{}
	err := binder.Bind(nacosProperties, NacosConfigPropertiesPrefix)
	return nacosProperties, err
}

// EnableNacosConfig 如果配置了对应的属性,则取配置的值;未配置,默认不开启nacos的配置加载
// func EnableNacosConfig() bool {
// 	return cast.ToBool(conf.Env.GetProperty("nacos.enable"))
// }

// NewNacosServerConfig 初始化nacos server配置
func NewNacosServerConfig(p *NacosConfigProperties) ([]constant.ServerConfig, error) {
	if !p.EnableConfig && !p.EnableNaming {
		return nil, nil
	}

	adds := p.ServerAddr
	if adds == "" {
		return nil, fmt.Errorf("nacos server addr is empty")
	}
	ipports, err := parseIPPort(adds)
	if err != nil {
		return nil, err
	}

	schema := p.Schema
	contextPath := p.ContextPath

	opts := []constant.ServerOption{}
	if schema != "" {
		opts = append(opts, constant.WithScheme(schema))
	}
	if contextPath != "" {
		opts = append(opts, constant.WithContextPath(contextPath))
	}

	sc := []constant.ServerConfig{}

	for _, ipport := range ipports {
		sc = append(sc, *constant.NewServerConfig(ipport.ip, ipport.port, opts...))
	}

	return sc, nil
}

// NewNacosClientConfig 初始化nacos client配置
func NewNacosClientConfig(p *NacosConfigProperties) (*constant.ClientConfig, error) {

	if p == nil || (!p.EnableConfig && !p.EnableNaming) {
		return nil, nil
	}

	namespace := p.NameSpace
	username := p.UserName
	password := p.Password

	opts := []constant.ClientOption{}

	// namespace
	if namespace != "" {
		opts = append(opts, constant.WithNamespaceId(namespace))
	}
	// username
	if username != "" {
		opts = append(opts, constant.WithUsername(username))
	}
	// password
	if password != "" {
		opts = append(opts, constant.WithPassword(password))
	}

	// TODO 其他配置

	clientConfig := constant.NewClientConfig(opts...)

	return clientConfig, nil
}

// NewNacosNamingClient 初始化nacos client
func NewNacosNamingClient(sc []constant.ServerConfig, cc *constant.ClientConfig, p *NacosConfigProperties) (naming_client.INamingClient, error) {
	if p == nil || !p.EnableNaming {
		return nil, nil
	}

	if len(sc) == 0 {
		return nil, errors.New("nacos server config is empty")
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

// NewNacosConfigClient 构建nacos config客户端
func NewNacosConfigClient(sc []constant.ServerConfig, cc *constant.ClientConfig, p *NacosConfigProperties) (config_client.IConfigClient, error) {
	if p == nil || !p.EnableConfig {
		return nil, nil
	}

	if len(sc) == 0 {
		return nil, errors.New("nacos server config is empty")
	}

	if cc == nil {
		return nil, errors.New("nacos client config is empty")
	}

	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  cc,
			ServerConfigs: sc,
		},
	)

	if err != nil {
		return nil, err
	}
	return configClient, nil
}

// NewNacosRemoteConfig 根据dataid获取nacos的远程配置
// func NewNacosRemoteConfig(env ag_conf.IConfigurableEnvironment, iClient config_client.IConfigClient, p *NacosConfigProperties) (*AutoconfigNacosConfig, error) {
func NewNacosRemoteConfig(env ag_conf.IConfigurableEnvironment, iClient config_client.IConfigClient, p *NacosConfigProperties) error {
	if p == nil || !p.EnableConfig {
		return nil
	}

	dateids := p.DataIDs

	// if len(NacosDataIDList.DataIDList) >= 1 {
	if len(dateids) >= 1 {

		// for _, dataidinfo := range NacosDataIDList.DataIDList {
		for _, dataidinfo := range dateids {
			if dataidinfo.DataID == "" {
				return fmt.Errorf("must config dataid value")
			}
			if dataidinfo.Group == "" {
				return fmt.Errorf("must config group value")
			}
			// 获取对应的nacos信息
			var context string
			context, err := iClient.GetConfig(vo.ConfigParam{
				DataId:  dataidinfo.DataID,
				Group:   dataidinfo.Group,
				Content: context,
				Type:    vo.YAML,
			})
			if err != nil {
				return fmt.Errorf("dataId:%s Group:%s parse error: %w", dataidinfo.DataID, dataidinfo.Group, err)
			}

			if context == "" {
				continue
			}
			err = addOrRefresh(env, context, &dataidinfo, false)
			if err != nil {
				return err
			}

			// 只要获取nacos的内容不返回error，就可以添加对应的监听
			iClient.ListenConfig(vo.ConfigParam{
				DataId: dataidinfo.DataID,
				Group:  dataidinfo.Group,
				Type:   vo.ConfigType(dataidinfo.Type), // 不指定类型能拿到吗
				OnChange: func(namespace string, group string, dataId string, data string) {
					// TODO dataId 和 group 是否可能不一致？
					err := addOrRefresh(env, data, &dataidinfo, true)
					if err != nil {
						slog.Error("nacos conf refresh", "dataId:", dataId, " errormsg:", err.Error())
					}
				},
			})
			// TODO 怎么取消配置监听
		}
	}
	return nil
}

// addOrRefresh或者刷新配置信息
// func addOrRefresh(env conf.IConfigurableEnvironment, context string, dataidinfo *model.DataIDInfo, refresh bool) error {
func addOrRefresh(env ag_conf.IConfigurableEnvironment, context string, dataidinfo *DataIDInfo, refresh bool) error {

	keyname := fmt.Sprintf("%s,%s", dataidinfo.DataID, dataidinfo.Group)
	res, err := parseContextByType(dataidinfo.Type, context)
	if err != nil {
		return err
	}
	if refresh {
		err := env.GetPropertySources().Replace(keyname, NewNacosPropertySource(keyname, res))
		return err
	}
	// 对于扩展的配置就是要在原来配置的后面添加
	env.GetPropertySources().AddLast(NewNacosPropertySource(keyname, res))

	return nil
}

// 解析从nacos上获取的内容
func parseContextByType(fileType string, context string) (map[string]any, error) {

	switch fileType {
	case "yaml":
		contextMap, err := yaml.Read([]byte(context))
		if err != nil {
			return nil, err
		}
		return ag_ext.GetFlattenedMap(contextMap)
	case "yml":
		contextMap, err := yaml.Read([]byte(context))
		// 此处是否要panic 要区分启动和日常刷新?
		if err != nil {
			return nil, err
		}
		return ag_ext.GetFlattenedMap(contextMap)
	case "json":
		contextMap, err := json.Read([]byte(context))
		if err != nil {
			return nil, err
		}
		return ag_ext.GetFlattenedMap(contextMap)
	case "properties":
		contextMap, err := prop.Read([]byte(context))
		if err != nil {
			return nil, err
		}
		return ag_ext.GetFlattenedMap(contextMap)
	default:
		return nil, fmt.Errorf("fileType:%s not be supported", fileType)
	}

}

type ipPort struct {
	ip   string
	port uint64
}

// parseIPPort 解析IP:端口格式的字符串，返回IP和端口的集合
func parseIPPort(input string) ([]ipPort, error) {
	var result []ipPort

	// 按逗号分割输入字符串
	entries := strings.Split(input, ",")

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		// 分割IP和端口
		parts := strings.Split(entry, ":")

		// 验证IP格式
		ip := net.ParseIP(parts[0])
		if ip == nil {
			return nil, fmt.Errorf("无效的IP地址: %s", parts[0])
		}

		// 处理端口
		var port uint64
		switch len(parts) {
		case 1:
			// 如果没有指定端口，使用默认端口80
			port = 8848
		case 2:
			// 验证端口格式
			portstr := parts[1]
			var err error
			if _, err = net.LookupPort("tcp", portstr); err != nil {
				return nil, fmt.Errorf("无效的端口: %s", portstr)
			}
			if port, err = cast.ToUint64E(portstr); err != nil {
				return nil, fmt.Errorf("无效的端口: %s", portstr)
			}
		default:
			return nil, errors.New("格式错误，应为IP:端口或IP")
		}

		result = append(result, ipPort{
			ip:   ip.String(),
			port: port,
		})
	}

	return result, nil
}
