package main

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"
)

//go:embed grpcTemplate.tpl
var grpcTemplate string

func (s *PackageInfo) execute() string {
	buf := new(bytes.Buffer)
	tmpl, err := template.New("grpc").Funcs(Funcs).Parse(strings.TrimSpace(grpcTemplate))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return strings.Trim(buf.String(), "\r\n")
}
