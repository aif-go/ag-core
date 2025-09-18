package protoc

import (
	"ag-core/tool/aggen/types"
	"ag-core/tool/aggen/util"
	"context"
	"fmt"
	"log/slog"
	"path"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	ModuleInfoKey       = "ModuleInfo"
	GlobalInfoKey       = "GlobalInfo"
	PkgInfoKey          = "PkgInfo"
	PackageGroupInfoKey = "PackageGroupInfo"
	PackageInfoKey      = "PackageInfo"
	ServiceInfoKey      = "ServiceInfo"
)

func Parse(gen *protogen.Plugin, files []*protogen.File) (*types.GennerInfo, error) {
	ctx := context.Background()
	// 获取模块信息
	moduleInfo, err := util.GetPwdModInfo()
	if err != nil {
		return nil, err
	}

	globalInfo := &types.GlobalInfo{
		ModuleInfo: moduleInfo,
	}

	ctx = context.WithValue(ctx, ModuleInfoKey, moduleInfo)
	ctx = context.WithValue(ctx, GlobalInfoKey, globalInfo)

	// 遍历proto文件解析
	for _, f := range files {
		if !f.Generate {
			slog.Warn(fmt.Sprintf("file %s not generate", f.GeneratedFilenamePrefix))
			continue
		}

		// 解析proto文件
		err = parseFile(ctx, globalInfo, f)
		if err != nil {
			return nil, err
		}
	}

	// 返回数据生成数据载体
	gennerInfo := &types.GennerInfo{
		ModuleInfo: globalInfo.ModuleInfo,
		GlobalInfo: globalInfo,
		// Versions:   make(map[string]string),
	}

	// 添加protoc版本信息
	protocVersion := protocVersion(gen)
	gennerInfo.SetBaseVersion("protoc", protocVersion)

	testLog("GlobalParse ========gennerInfo==========", gennerInfo)
	return gennerInfo, nil
}

// parseFile 解析proto文件
func parseFile(ctx context.Context, gInfo *types.GlobalInfo, f *protogen.File) error {
	pkg, err := parsePackageInfo(ctx, f)
	if err != nil {
		return err
	}

	pkg.ModuleInfo = gInfo.ModuleInfo

	err = gInfo.AddPackageInfo(pkg)
	if err != nil {
		return err
	}

	return nil
}

// parsePackageInfo 解析package信息
func parsePackageInfo(ctx context.Context, f *protogen.File) (*types.PackageInfo, error) {
	pkg := &types.PackageInfo{}

	pkg.NoFastAPI = true // 原kitex保留，兼容模板

	pkg.Codec = "protobuf"
	pkg.IDLName = f.Desc.Path()

	// PROTO: option go_package
	pth := strings.Trim(string(f.GoImportPath), "\"")
	if pth == "" {
		return nil, fmt.Errorf("missing %q option in %q", "go_package", f.Desc.FullName())
	}

	pi := types.PkgInfo{
		PkgName:    f.Proto.GetPackage(), // PROTO: package TODO 不可为空吧？
		ImportPkg:  strings.ReplaceAll(f.Proto.GetPackage(), ".", "/"),
		PkgRefName: goSanitized(path.Base(pth)),
		ImportPath: pth,
	}
	pkg.PkgInfo = pi

	ctx = context.WithValue(ctx, PkgInfoKey, pi)
	ctx = context.WithValue(ctx, PackageInfoKey, pkg)

	// testLog("parsePackageInfo", pi)

	ss := make([]*types.ServiceInfo, 0)

	for _, s := range f.Services {
		// 解析service信息
		si, err := parseServiceInfo(ctx, s)
		if err != nil {
			return nil, err
		}

		si.PkgInfo = pi // 冗余

		if f.Generate { // FIXME main顶层就过滤掉了，到这里Generate均为true
			si.GenerateHandler = true
		}

		si.PkgInfo = pi // 冗余
		ss = append(ss, si)
	}

	pkg.Services = ss
	return pkg, nil
}

// parseServiceInfo 解析service信息
func parseServiceInfo(ctx context.Context, s *protogen.Service) (*types.ServiceInfo, error) {
	si := &types.ServiceInfo{
		ServiceName:    s.GoName,
		RawServiceName: string(s.Desc.Name()),

		Deprected: s.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated(),
		Comment:   convertComment(s.Comments),
	}
	si.ServiceTypeName = func() string { return si.PkgRefName + "." + si.ServiceName }

	ctx = context.WithValue(ctx, ServiceInfoKey, si)

	methods := s.Methods
	// 解析method信息
	for _, m := range methods {
		mi, err := parseMethodInfo(ctx, m)
		if err != nil {
			return nil, err
		}

		si.Methods = append(si.Methods, mi)
		if mi.IsStreaming {
			si.HasStreaming = true
		}
	}

	return si, nil
}

// parseMethodInfo 解析method信息
func parseMethodInfo(ctx context.Context, m *protogen.Method) (*types.MethodInfo, error) {

	req := convertParameter(m.Input, "Req")
	res := convertParameter(m.Output, "Resp")

	// 解析Http options
	hds, err := parserMethodHttpDesc(m)
	if err != nil {
		return nil, err
	}

	pkgInfo := ctx.Value(PkgInfoKey).(types.PkgInfo)
	si := ctx.Value(ServiceInfoKey).(*types.ServiceInfo)

	methodName := m.GoName
	mi := &types.MethodInfo{
		PkgInfo:            pkgInfo,
		ServiceName:        si.ServiceName,
		RawName:            string(m.Desc.Name()),
		Name:               methodName,
		Args:               []*types.Parameter{req}, // FIXME 暂不支持多个请求参数
		Resp:               res,
		ArgStructName:      methodName + "Args",
		ResStructName:      methodName + "Result",
		GenArgResultStruct: true, // TODO 作用是什么
		ClientStreaming:    m.Desc.IsStreamingClient(),
		ServerStreaming:    m.Desc.IsStreamingServer(),

		Deprected: m.Desc.Options().(*descriptorpb.MethodOptions).GetDeprecated(),
		Comment:   convertComment(m.Comments),
		HttpDescs: hds,
	}

	mi.ArgsLength = len(mi.Args) // TODO proto场景下只有一个

	if mi.ClientStreaming || mi.ServerStreaming {
		mi.IsStreaming = true
	}
	return mi, nil
}

// ConvertParameter converts a protogen.Parameter to a types.Parameter.
func convertParameter(msg *protogen.Message, paramName string) *types.Parameter {
	importPath := util.FixImport(msg.GoIdent.GoImportPath.String())
	pkgRefName := util.GoSanitized(path.Base(importPath))
	res := &types.Parameter{
		Deps: []types.PkgInfo{
			{
				PkgRefName: pkgRefName,
				ImportPath: importPath,
			},
		},
		Name:      paramName,
		RawName:   paramName,
		Type:      "*" + pkgRefName + "." + msg.GoIdent.GoName,
		UnptrType: pkgRefName + "." + msg.GoIdent.GoName,
		RawType:   msg.GoIdent.GoName,
	}
	return res
}

// convertComment 解析注释
func convertComment(pcm protogen.CommentSet) types.CommentSet {
	cm := types.CommentSet{
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
