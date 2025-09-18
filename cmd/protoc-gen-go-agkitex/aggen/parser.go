package aggen

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"path"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
)

// 解析PackageInfo
func ParsePackageInfo_protoc(gen *protogen.Plugin, f *protogen.File) *PackageInfo {

	// TODO file options
	_fopts := f.Desc.Options()
	slog.Info("file:", "optstype", fmt.Sprintf("%T", _fopts), "opts", _fopts)
	fopts, _ := _fopts.(*descriptorpb.FileOptions)
	fopts_json, _ := json.MarshalIndent(fopts, "", "  ")
	slog.Info(fmt.Sprintf("fopts:%s", fopts_json))

	pkg := &PackageInfo{}
	/*
		1.package 信息
	*/
	pkg.ProtocVersion = protocVersion(gen)

	// NoFastAPI
	pkg.NoFastAPI = true
	// Codec
	pkg.Codec = "protobuf" // protoc方式默认使用protobuf编码
	// IDLName
	pkg.IDLName = f.Desc.Path()
	// pkg.IDLName, _ = filepath.Abs(pkg.IDLName)

	pth := strings.Trim(string(f.GoImportPath), "\"")
	pi := PkgInfo{
		PkgName:    f.Proto.GetPackage(),
		PkgRefName: goSanitized(path.Base(pth)),
		ImportPath: pth,
	}
	/* 3.convertTypes */
	// 遍历services
	ss := make([]*ServiceInfo, 0)
	for _, s := range f.Services {
		si := &ServiceInfo{
			PkgInfo:        pi,
			ServiceName:    s.GoName,
			RawServiceName: string(s.Desc.Name()),
		}
		si.ServiceTypeName = func() string { return si.PkgRefName + "." + si.ServiceName }

		slog.Info("---------------------------")
		slog.Info("service:", "ServiceName", si.ServiceName)
		slog.Info("service:", "RawServiceName", si.RawServiceName)
		slog.Info("service:", "ServiceTypeName", si.ServiceTypeName())

		// TODO service options
		_sopts := s.Desc.Options()
		slog.Info("service:", "optstype", fmt.Sprintf("%T", _sopts), "opts", _sopts)
		sopts, _ := _sopts.(*descriptorpb.ServiceOptions)
		sopts_json, err := json.MarshalIndent(sopts, "", "  ")
		slog.Info(fmt.Sprintf("sopts: %s, err: %v", sopts_json, err))

		// 遍历方法
		for _, m := range s.Methods {
			slog.Info("> method:", "name", m.GoName)

			// TODO method options
			_mopts := m.Desc.Options()
			slog.Info("method:", "optstype", fmt.Sprintf("%T", _mopts), "opts", _mopts)
			mopts, ok := _mopts.(*descriptorpb.MethodOptions)
			// mopts, _ := _mopts.(*descriptorpb.ServiceOptions)
			mopts_json, err := json.MarshalIndent(mopts, "", "  ")
			slog.Info(fmt.Sprintf("mopts: %s, ok: %v err: %v", mopts_json, ok, err))

			req := convertParameter(m.Input, "Req")
			res := convertParameter(m.Output, "Resp")

			// reqjson, _ := json.MarshalIndent(req, "", "  ")
			// resjson, _ := json.MarshalIndent(res, "", "  ")
			// slog.Info(fmt.Sprintf("> method:%s req:%s", m.GoName, reqjson))
			// slog.Info(fmt.Sprintf("> method:%s res:%s", m.GoName, resjson))

			methodName := m.GoName
			mi := &MethodInfo{
				PkgInfo:            pi,
				ServiceName:        si.ServiceName,
				RawName:            string(m.Desc.Name()),
				Name:               methodName,
				Args:               []*Parameter{req},
				Resp:               res,
				ArgStructName:      methodName + "Args",
				ResStructName:      methodName + "Result",
				GenArgResultStruct: true, // TODO 作用是什么
				ClientStreaming:    m.Desc.IsStreamingClient(),
				ServerStreaming:    m.Desc.IsStreamingServer(),
			}
			si.Methods = append(si.Methods, mi)
			if mi.ClientStreaming || mi.ServerStreaming {
				mi.IsStreaming = true
				si.HasStreaming = true
			}
		}

		if f.Generate { // FIXME main顶层就过滤掉了，到这里Generate均为true
			si.GenerateHandler = true
		}
		ss = append(ss, si)
	}

	// TODO CombineServices

	pkg.Services = ss
	return pkg
}
