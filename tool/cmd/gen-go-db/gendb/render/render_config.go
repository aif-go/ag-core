package render

var MY_SQL, DB2 string = "mysql", "db2"
var EntityPathSuffix, DaoPathSuffix, SqlPathSuffix, YamlPathSuffix string = "repository/model/", "repository/dao/", "repository/ddl/", "repository/yaml/"

// BaseConfig 基础配置

type BaseConfig struct {
	DbTemplatePath    string // DB模板文件路径
	PackageNamePrefix string // Go模块前缀
	DbType            string // 数据库类型
	OutputPath        string // 输出路径
}

// GenerateOptions 生成选项

type GenerateOptions struct {
	Entityable bool // 是否生成实体
	DDLable    bool // 是否生成DDL
	Daoable    bool // 是否生成DAO
	Sqlable    bool // 是否生成SQL
}

// PathConfig 路径配置

type PathConfig struct {
	EntityPath string // 实体文件路径
	DaoPath    string // DAO文件路径
	SqlPath    string // SQL文件路径
	YamlPath   string // YAML文件路径
}

// SupportConfig 支持配置

type SupportConfig struct {
	SupportDB     []string          // 支持的数据库类型
	SupportTables map[string]string // 指定需要生成的表
}

// AGInfraStructrueConfig 综合配置结构体

type AGInfraStructrueConfig struct {
	BaseConfig      // 基础配置
	GenerateOptions // 生成选项
	PathConfig      // 路径配置
	SupportConfig   // 支持配置
	// Encode            string
	// Sort              string
	// Engine            string
}
