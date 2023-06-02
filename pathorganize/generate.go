// pathorganize is change the file generation pathes
// Opeanpi files placed into ../api/openapi directory
// Cmd files placed into ../cmd/ xx directory and added goa- prefix for the dirs
package pathorganize

import (
	"fmt"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
)

// Register the plugin Generator functions.
func init() {
	codegen.RegisterPluginLast("pathmod", "gen", nil, Generate)
	codegen.RegisterPluginLast("pathmod-example", "example", nil, UpdateExample)
}

var needReplace = true

func ReplaceGen(s string) (res string) {
	if needReplace {
		res = strings.Replace(s, "gen/gen/", "gen/", -1)
		res = strings.Replace(res, "gen\\gen\\", "gen\\", -1)
		return
	}
	res = s
	return
}

// Generate is rewrite generated files path
func Generate(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
	for _, f := range files {

		f.Path = ReplaceGen(f.Path)

		// rewrite openapi output path
		if strings.Contains(f.Path, "http\\openapi") || strings.Contains(f.Path, "http/openapi") {
			fn := filepath.Base(f.Path)
			f.Path = fmt.Sprintf("../api/openapi/%s", fn)
		}

		for _, s := range f.SectionTemplates {
			hd, ok := s.Data.(map[string]interface{})
			if !ok {
				continue
			}
			specs, ok := hd["Imports"].([]*codegen.ImportSpec)
			if !ok {
				continue
			}
			for _, is := range specs {
				is.Path = ReplaceGen(is.Path)
			}
		}
	}
	return files, nil
}

// UpdateExample is update example files path
func UpdateExample(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
	for _, f := range files {
		// rewrite base path
		f.Path = ReplaceGen(f.Path)
		if strings.Contains(f.Path, "cmd\\") || strings.Contains(f.Path, "cmd/") {
			f.Path = strings.Replace(f.Path, "cmd\\", "..\\cmd\\goa-", -1)
			f.Path = strings.Replace(f.Path, "cmd/", "../cmd/goa-", -1)
		}

		// rewrite implementation path
		isSvc := false
		if strings.Contains(f.Path, "\\") == false && strings.Contains(f.Path, "/") == false {
			fn := filepath.Base(f.Path)
			f.Path = fmt.Sprintf("../endpoint/worker/%s", fn)
			isSvc = true
		}

		for _, s := range f.SectionTemplates {
			hd, ok := s.Data.(map[string]interface{})
			if !ok {
				continue
			}
			specs, ok := hd["Imports"].([]*codegen.ImportSpec)
			if !ok {
				continue
			}
			for _, is := range specs {
				is.Path = ReplaceGen(is.Path)
			}
			if isSvc {
				hd["Pkg"] = "worker"
			}
		}
	}
	return files, nil
}
