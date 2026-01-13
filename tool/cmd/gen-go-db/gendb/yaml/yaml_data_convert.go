package yaml

import "ag-core/tool/cmd/gen-go-db/gendb/render"

type YamlDataConvert struct {
}

// Convert 将源文件数据转换为模板文件数据
func (convert *YamlDataConvert) Convert(yamlDataList []*YamlDataConfig) []*render.TableData {

	list := []*render.TableData{}
	for _, yamlData := range yamlDataList {
		renderTableData:=ConvertConfigToRenderTableData(*yamlData)
		list = append(list, renderTableData)
	}
	return list
}