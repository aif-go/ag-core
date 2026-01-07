package erm

import "ag-core/tool/cmd/gen-go-db/gendb/render"

type DataConvert[T any] interface {
	Convert(srcData T) *render.TableData
}