package muxer

/*
import (
	"testing"

	"goa.design/goa/v3/codegen/generator"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestGenerate(t *testing.T) {
	cases := map[string]struct {
		DSL      func()
		ExpFiles int
	}{
		"multi-simple": {testdata.MultiSimpleDSL, 1},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			codegen.RunHTTPDSL(t, c.DSL)
			roots := []eval.Root{expr.Root}
			files := generateFiles(t, roots)
			newFiles, err := Generate("", roots, files)
			if err != nil {
				t.Fatalf("generate error: %v", err)
			}
			newFilesCount := len(newFiles) - len(files)
			if newFilesCount != c.ExpFiles {
				t.Errorf("invalid code: number of new files expected %d, got %d", c.ExpFiles, newFilesCount)
			}
		})
	}
}

func generateFiles(t *testing.T, roots []eval.Root) []*codegen.File {
	files, err := generator.Service("", roots)
	if err != nil {
		t.Fatalf("error in code generation: %v", err)
	}
	httpFiles, err := generator.Transport("", roots)
	if err != nil {
		t.Fatalf("error in HTTP code generation: %v", err)
	}
	files = append(files, httpFiles...)
	return files
}

func generateExamples(t *testing.T, roots []eval.Root) []*codegen.File {
	files, err := generator.Example("", roots)
	if err != nil {
		t.Fatalf("error in code generation: %v", err)
	}
	return files
}
*/
