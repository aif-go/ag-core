package main

import (
	_ "embed"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
)

//go:embed serverTemplate.tbl
var serverTemplate string // 服务接口模板

type protoFile struct {
	DescPath                string
	GeneratedFilenamePrefix string
	Deprected               bool
	GoImportPath            string
	GoPackageName           string
	ServiceDescs            []*serviceDesc
}

type serviceDesc struct {
	GoName         string
	FullName       string
	ProtoDeprected bool
	Deprected      bool
	Methods        []*methodDesc
	Comment        commentSet
}

type methodDesc struct {
	GoName                string
	DescName              string
	IsStreamingClient     bool
	IsStreamingServer     bool
	InputGoIdent          protogen.GoIdent
	InputQualifiedString  string
	OutputGoIdent         protogen.GoIdent
	OutputQualifiedString string

	Comment commentSet
}

type commentSet struct {
	LeadingDetacheds []string
	Leading          string
	Trailing         string
}

// parseProtoFile 解析 .proto 信息
func parseProtoFile(g *protogen.GeneratedFile, file *protogen.File) *protoFile {
	pf := &protoFile{
		DescPath:                file.Desc.Path(),
		GeneratedFilenamePrefix: file.GeneratedFilenamePrefix,
		Deprected:               file.Proto.GetOptions().GetDeprecated(),
		GoImportPath:            string(file.GoImportPath),
		GoPackageName:           string(file.GoPackageName),
	}

	for _, s := range file.Services {
		sd := &serviceDesc{
			GoName:         s.GoName,
			FullName:       string(s.Desc.FullName()),
			ProtoDeprected: pf.Deprected,
			Deprected:      s.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated(),
			Comment:        newComment(s.Comments),
		}
		for _, m := range s.Methods {
			md := &methodDesc{
				GoName:                m.GoName,
				DescName:              string(m.Desc.Name()),
				IsStreamingClient:     m.Desc.IsStreamingClient(),
				IsStreamingServer:     m.Desc.IsStreamingServer(),
				InputGoIdent:          m.Input.GoIdent,
				InputQualifiedString:  g.QualifiedGoIdent(m.Input.GoIdent),
				OutputGoIdent:         m.Output.GoIdent,
				OutputQualifiedString: g.QualifiedGoIdent(m.Output.GoIdent),
				Comment:               newComment(m.Comments),
			}

			sd.Methods = append(sd.Methods, md)
		}

		pf.ServiceDescs = append(pf.ServiceDescs, sd)
	}
	return pf
}

// newComment 解析注释
func newComment(pcm protogen.CommentSet) commentSet {
	cm := commentSet{
		Leading:  trimeComment(pcm.Leading.String()),
		Trailing: trimeComment(pcm.Trailing.String()),
	}
	for _, pld := range pcm.LeadingDetached {
		cm.LeadingDetacheds = append(cm.LeadingDetacheds, trimeComment(pld.String()))
	}
	return cm
}

func trimeComment(comment string) string {
	if comment == "" {
		return ""
	}
	return strings.Trim(strings.TrimPrefix(strings.TrimSuffix(comment, "\n"), "//"), " ")
}
