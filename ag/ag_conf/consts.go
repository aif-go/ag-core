package ag_conf

import "fmt"

// CommandArgsPrefix 命令行参数前缀取值key
const CommandArgsPrefix = "GS_ARGS_PREFIX"

const (
	//ConstPlaceholderPrefix Prefix for system property placeholders: "${".
	ConstPlaceholderPrefix = "${"
	//ConstPlaceholderSuffix Suffix for system property placeholders: "}".
	ConstPlaceholderSuffix = "}"
	//ConstValueSeparator Value separator for system property placeholders: ":".
	ConstValueSeparator = ":"
	// ConstEncryptKeyWords value start with keywords means has been encrypt
	ConstEncryptKeyWords = "{cipher}"

	SourceKeySysPrefix     string = "[SYS]"
	SourceKeyLocalPrefix   string = "[LOCAL]"
	SourceKeyDecryptPrefix string = "[DECRYPT]"

// // CustomerYamlSource 客户配置的yaml资源
// CustomerYamlSource string = "CustomerYamlSource"
// // CustomerPropertiesSource 客户配置的properties资源
// CustomerPropertiesSource string = "CustomerPropertiesSource"
// // LocalDefaultYamlSource 二进制可执行文件中默认配置的yaml文件
// LocalDefaultYamlSource string = "LocalDefaultYamlSource"
// // LocalDefaultPropertiesSource 二进制可执行中默认配置的properties文件
// LocalDefaultPropertiesSource string = "LocalDefaultPropertiesSource"
// // LocalProfileYamlSource 基于用户配置的profile加载对应的yaml文件
// LocalProfileYamlSource string = "LocalProfileYamlSource"
// // LocalProfilePropertiesSource 基于用户配置的profile加载对应的properties文件
// LocalProfilePropertiesSource string = "LocalProfilePropertiesSource"

)

var (
	SourceKeySystemProperties  = fmt.Sprintf("%s-%s", SourceKeySysPrefix, "Properties")
	SourceKeySystemEnvironment = fmt.Sprintf("%s-%s", SourceKeySysPrefix, "Environment")
	SourceKeyDecryptSystem     = fmt.Sprintf("%s-%s", SourceKeyDecryptPrefix, "SystemEnvAndProp")
	SourceKeyDecryptLocal      = fmt.Sprintf("%s-%s", SourceKeyDecryptPrefix, SourceKeyLocalPrefix)
)
