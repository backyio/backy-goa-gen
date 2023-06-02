// micromuxer is generate custom service handler for Go-Micro
package micromuxer

import (
	"fmt"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type (
	serviceData struct {
		ServerPath      string
		HttpServerPath  string
		ServerAlias     string
		HttpServerAlias string
		NewServer       string
	}
	services = []serviceData
)

var (
	middlewarePath string
)

// getMiddlewarePath generate path for middleware from pkg name
func getMiddlewarePath(genpkg string) string {
	p := filepath.Join(genpkg, "middleware")
	p = strings.Replace(p, "\\", "/", -1)
	p = strings.Replace(p, "/gen", "", -1)
	return p
}

// Register the plugin Generator functions.
func init() {
	codegen.RegisterPluginFirst("micro-muxer", "example", nil, Generate)
}

// Generate generates go-muxer specific file.
func Generate(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
	middlewarePath = getMiddlewarePath(genpkg)
	var svcs services
	for _, root := range roots {
		if r, ok := root.(*expr.RootExpr); ok {
			svcs = append(svcs, CollectServices(genpkg, r)...)
		}
	}

	return append(files, GenerateMicroMuxerFile(genpkg, svcs)), nil
}

func RepPath(p string) string {
	return strings.Replace(p, "\\", "/", -1)
}

// CollectServices collecting information about all services
func CollectServices(genpkg string, root eval.Root) (data []serviceData) {
	scope := codegen.NewNameScope()
	if r, ok := root.(*expr.RootExpr); ok {
		// Add the generated main files
		for _, svc := range r.API.HTTP.Services {
			data = append(data, serviceData{
				ServerPath:      RepPath(filepath.Join(genpkg, codegen.Gendir, (svc.Name()))),
				ServerAlias:     svc.Name(),
				HttpServerPath:  RepPath(filepath.Join(genpkg, codegen.Gendir, "http", (svc.Name()), "server")),
				HttpServerAlias: fmt.Sprintf("%s%s", scope.Unique(svc.Name()), "srv"),
				NewServer:       fmt.Sprintf("New%s", codegen.CamelCase(svc.Name(), true, false)),
			})
		}
	}
	return
}

// GenerateMicroMuxerFile returns the generated go muxer file.
func GenerateMicroMuxerFile(genpkg string, svc services) *codegen.File {
	path := "micro.go"
	title := "Go-Micro muxer generator"

	imp := []*codegen.ImportSpec{}
	imp = append(imp, []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "net/http"},
		{Path: "go-micro.dev/v4/logger", Name: "mlog"},
		{Path: middlewarePath},
		{Path: "goa.design/goa/v3/http", Name: "goahttp"},
	}...)

	for _, v := range svc {
		imp = append(imp, []*codegen.ImportSpec{
			{Path: v.ServerPath, Name: v.ServerAlias},
			{Path: v.HttpServerPath, Name: v.HttpServerAlias},
		}...)
	}

	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "service", imp),
	}

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "go-micro-muxer",
		Data:   map[string]interface{}{"services": svc},
		Source: muxerT,
	})

	return &codegen.File{Path: path, SectionTemplates: sections}
}

const muxerT = `

type (
	MicroHttpEndpoint interface {
		Service() string
		MethodNames() []string
		Mount(goahttp.Muxer)
		Use(func(http.Handler) http.Handler)
	}

	MicroHttpEndpoints map[string]MicroHttpEndpoint
)

var AvailableHttpEndpoints MicroHttpEndpoints = make(MicroHttpEndpoints)

// NewMicroMuxer initialize the services and returns http handler
func NewMicroMuxer(l mlog.Logger, enabled map[string]bool) (http.Handler, goahttp.MiddlewareMuxer) {
	var (
		eh      = middleware.ErrorHandler(l)
		dec     = goahttp.RequestDecoder
		enc     = goahttp.ResponseEncoder
		mux     = goahttp.NewMuxer()
	)

	{{- range .services }}
	{
		if b, ok := enabled[{{ .ServerAlias }}.ServiceName]; len(enabled) == 0 || ok && b {
			{{ .ServerAlias }}Svc := {{ .NewServer }}(l)
			{{ .ServerAlias }}Endpoints := {{ .ServerAlias }}.NewEndpoints({{ .ServerAlias }}Svc)
			{{ .ServerAlias }}Server := {{ .HttpServerAlias }}.New({{ .ServerAlias }}Endpoints, mux, dec, enc, eh, nil)
			{{ .HttpServerAlias }}.Mount(mux, {{ .ServerAlias }}Server)
            AvailableHttpEndpoints[{{ .ServerAlias }}Server.Service()] = {{ .ServerAlias }}Server
		}
	}
	{{- end }}

	return http.Handler(mux), mux
}
`
