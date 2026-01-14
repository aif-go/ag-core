package generate

import "ag-core/tool/cmd/gen-go-db/gendb/render"

// Generate 提供资源文件的解析行为
type Generate[T any] interface {
	ParseTemplateFile(config *render.AGInfraStructrueConfig) (*T, error)
	Generate(config *render.AGInfraStructrueConfig, srcData *T) error
	// GenerateEnity(config AGInfraStructrueConfig,srcData *T)
	// GenerateDao(config AGInfraStructrueConfig,srcData *T)
}
