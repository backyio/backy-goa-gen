package muxer

/*
import (
	"path"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	httpcode "goa.design/goa/v3/http/codegen"
)

type (
	httpService struct {
		Data   *service.Data
		ApiPkg string
	}
	httpServiceExpr struct {
		Data   *expr.HTTPServiceExpr
		ApiPkg string
	}
	muxData struct {
		Services    []*httpServiceExpr
		ServiceData []*httpService
		File        *codegen.File
	}
)

func init() {
	codegen.RegisterPluginLast("backy-muxer", "example", nil, Generate)
}

// Generate produces muxer code that handle preflight requests and updates
// the HTTP responses with the appropriate CORS headers.
func Generate(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
	var data muxData
	genFiles := files
	for _, root := range roots {
		r, ok := root.(*expr.RootExpr)
		if !ok {
			continue // could be a plugin root expression
		}
		if r.API != nil && r.API.HTTP != nil && len(r.API.HTTP.Services) > 0 {
			for _, v := range r.API.HTTP.Services {
				data.Services = append(data.Services, &httpServiceExpr{
					Data:   v,
					ApiPkg: r.API.Name,
				})
			}
		}

	}
	GenMuxer(genpkg, &data)
	if data.File != nil {
		genFiles = append(genFiles, data.File)
	}
	return genFiles, nil
}

// GenMuxer is generate muxer specific service creation
func GenMuxer(genpkg string, mux *muxData) {
	var spec []*codegen.ImportSpec

	spec = append(spec, codegen.GoaImport("context"))
	spec = append(spec, codegen.GoaImport("log"))
	spec = append(spec, codegen.GoaImport("net/http"))
	spec = append(spec, codegen.GoaImport("net/url"))

	// Iterate through services listed in the muxer expression.
	scope := codegen.NewNameScope()
	for _, svc := range mux.Services {
		sd := httpcode.HTTPServices.Get(svc.Data.Name())
		spec = append(spec, &codegen.ImportSpec{
			Path: path.Join(genpkg, sd.Service.PathName),
			Name: scope.Unique(sd.Service.PkgName),
		})
		mux.ServiceData = append(mux.ServiceData, &httpService{
			Data:   sd.Service,
			ApiPkg: svc.ApiPkg,
		})
	}

	var (
		rootPath string
		apiPkg   string
	)
	{
		// genpkg is created by path.Join so the separator is / regardless of operating system
		idx := strings.LastIndex(genpkg, string("/"))
		rootPath = "."
		if idx > 0 {
			rootPath = genpkg[:idx]
		}
		apiPkg = scope.Unique(strings.ToLower(codegen.Goify(root.API.Name, false)), "api")
	}
	spec = append(spec, &codegen.ImportSpec{Path: rootPath, Name: apiPkg})

	path := filepath.Join(codegen.Gendir, "muxer.go")

	sections := []*codegen.SectionTemplate{
		codegen.Header("Backy.io muxer helper", "muxer", spec),
	}

	sections = append(sections, &codegen.SectionTemplate{
		Name:    "backy-logger",
		Source:  mainLoggerT,
		Data:    nil,
		FuncMap: nil,
	})

	fm := codegen.TemplateFuncs()
	fm["join"] = strings.Join
	sections = append(sections, &codegen.SectionTemplate{
		Name:   "backy-muxer",
		Source: muxerT,
		Data: map[string]any{
			"Services": mux.ServiceData,
		},
		FuncMap: fm,
	})
	//	mux.ServiceData[0].Service.PathName
	mux.File = &codegen.File{Path: path, SectionTemplates: sections}
}

const (
	mainLoggerT = `
	var (
		logger *log.Logger
	)
`
	muxerT = `
   {{ comment "Initialize the services." }}
	var (
	{{- range .Services }}
		{{- if .Service.Methods }}
		{{ .Service.VarName }}Svc {{ .Service.PkgName }}.Service
		{{- end }}
	{{- end }}
`
	mainSvcsT = `
	{{ comment "Initialize the services." }}
	var (
	{{- range .Services }}
		{{- if .Methods }}
		{{ .VarName }}Svc {{ .PkgName }}.Service
		{{- end }}
	{{- end }}
	)
	{
	{{- range .Services }}
		{{- if .Methods }}
		{{ .VarName }}Svc = {{ $.APIPkg }}.New{{ .StructName }}(logger)
		{{- end }}
	{{- end }}
	}
`
)
*/
