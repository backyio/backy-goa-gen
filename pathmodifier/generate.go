package pathmodifier

import (
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
)

// Register the plugin Generator functions.
func init() {
	codegen.RegisterPluginLast("pathmod", "gen", nil, Generate)
	codegen.RegisterPluginLast("pathmod-example", "example", nil, UpdateExample)
}

// Generate is rewrite generated files path
func Generate(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
	for _, f := range files {
		f.Path = strings.Replace(f.Path, "/gen/", "/", -1)
		for _, s := range f.SectionTemplates {
			hd := s.Data.(map[string]interface{})
			specs := hd["Imports"].([]*codegen.ImportSpec)
			for _, is := range specs {
				is.Path = strings.Replace(is.Path, "/gen/", "/", -1)
			}
		}
	}
	return files, nil
}

// UpdateExample is update example files path
func UpdateExample(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
	for _, f := range files {
		f.Path = strings.Replace(f.Path, "/gen/", "/", -1)
		for _, s := range f.SectionTemplates {
			hd := s.Data.(map[string]interface{})
			specs := hd["Imports"].([]*codegen.ImportSpec)
			for _, is := range specs {
				is.Path = strings.Replace(is.Path, "/gen/", "/", -1)
			}
		}
	}
	return files, nil
}
