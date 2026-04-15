package reader

import (
	"ag-core/ag/ag_conf/reader/json"
	"ag-core/ag/ag_conf/reader/prop"
	"ag-core/ag/ag_conf/reader/toml"
	"ag-core/ag/ag_conf/reader/yaml"
)

type Reader func(b []byte) (map[string]any, error)

var Readers = map[string]Reader{
	"yaml":       yaml.Read,
	"yml":        yaml.Read,
	"json":       json.Read,
	"properties": prop.Read,
	"toml":       toml.Read,
}
