package reader

import (
	"github.com/aif-go/ag-core/ag/ag_conf/reader/json"
	"github.com/aif-go/ag-core/ag/ag_conf/reader/prop"
	"github.com/aif-go/ag-core/ag/ag_conf/reader/toml"
	"github.com/aif-go/ag-core/ag/ag_conf/reader/yaml"
)

type Reader func(b []byte) (map[string]any, error)

var Readers = map[string]Reader{
	"yaml":       yaml.Read,
	"yml":        yaml.Read,
	"json":       json.Read,
	"properties": prop.Read,
	"toml":       toml.Read,
}
