package main

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"
)

//go:embed httpTemplate.tpl
var httpTemplate string

//go:embed serverTemplate.tbl
var serverTemplate string

type serviceDesc struct {
	ServiceType string // Greeter
	ServiceName string // helloworld.Greeter
	Metadata    string // api/helloworld/helloworld.proto
	Methods     []*methodDesc
	MethodSets  map[string]*methodDesc
}

type methodDesc struct {
	// method
	Name         string
	OriginalName string // The parsed original name
	Num          int
	Request      string
	Reply        string
	Comment      string
	// http_rule
	Path         string
	PathVars     []string
	Method       string
	HasVars      bool
	HasBody      bool
	Body         string
	ResponseBody string
}

func (s *serviceDesc) execute(t string) string {
	s.MethodSets = make(map[string]*methodDesc)
	for _, m := range s.Methods {
		s.MethodSets[m.Name] = m
	}
	buf := new(bytes.Buffer)
	var tpl string
	switch t {
	case "server":
		tpl = serverTemplate
	case "http":
		tpl = httpTemplate
	}
	println(t, buf.String())
	tmpl, err := template.New(tpl).Parse(strings.TrimSpace(tpl))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return strings.Trim(buf.String(), "\r\n")
}
