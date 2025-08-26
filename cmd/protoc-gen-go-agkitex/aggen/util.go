// Copyright 2021 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aggen

import (
	"fmt"
	"go/token"
	"path"
	"strings"
	"unicode"

	"google.golang.org/protobuf/compiler/protogen"
)

// goSanitized returns a valid Go identifier. 字符合法化
func goSanitized(s string) string {
	// replace invalid characters with '_'
	s = strings.Map(func(r rune) rune {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return '_'
		}
		return r
	}, s)

	// avoid invalid identifier
	if !token.IsIdentifier(s) {
		return "_" + s
	}
	return s
}

func convertParameter(msg *protogen.Message, paramName string) *Parameter {
	importPath := fixImport(msg.GoIdent.GoImportPath.String())
	pkgRefName := goSanitized(path.Base(importPath))
	res := &Parameter{
		Deps: []PkgInfo{
			{
				PkgRefName: pkgRefName,
				ImportPath: importPath,
			},
		},
		Name:    paramName,
		RawName: paramName,
		Type:    "*" + pkgRefName + "." + msg.GoIdent.GoName,
	}
	return res
}

func fixImport(path string) string {
	path = strings.Trim(path, "\"")
	return path
}

func protocVersion(gen *protogen.Plugin) string {
	v := gen.Request.GetCompilerVersion()
	if v == nil {
		return "(unknown)"
	}
	var suffix string
	if s := v.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", v.GetMajor(), v.GetMinor(), v.GetPatch(), suffix)
}

func NeedCallOpt(pkg *PackageInfo) bool {
	needCallOpt := false
	if pkg.StreamX {
		for _, m := range pkg.ServiceInfo.AllMethods() {
			if !(m.ClientStreaming || m.ServerStreaming) {
				needCallOpt = true
				break
			}
		}
		return needCallOpt
	} else {
		// callopt is referenced only by non-streaming methods
		switch pkg.Codec {
		case "thrift":
			for _, m := range pkg.ServiceInfo.AllMethods() {
				if !m.IsStreaming {
					needCallOpt = true
					break
				}
			}
		case "protobuf":
			needCallOpt = true
		}
		return needCallOpt
	}
}
