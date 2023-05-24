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

// Register the plugin Generator functions.
func init() {
	codegen.RegisterPluginFirst("micro-muxer", "example", nil, Generate)
}

// Generate generates go-muxer specific file.
func Generate(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
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
		{Path: "goa.design/goa/v3/middleware"},
		{Path: "goa.design/goa/v3/http/middleware", Name: "httpmdlwr"},
		{Path: "goa.design/goa/v3/http", Name: "goahttp"},
		{Path: RepPath(filepath.Join(genpkg, codegen.Gendir, "log")), Name: "log"},
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
// NewMicroMuxer initialize the services and returns http handler
func NewMicroMuxer(l mlog.Logger, enabled map[string]bool) (http.Handler, goahttp.MiddlewareMuxer) {
	logger := &log.Logger{l}
	var (
		adapter = logger
		eh      = errorHandler(logger)
		dec     = goahttp.RequestDecoder
		enc     = goahttp.ResponseEncoder
		mux     = goahttp.NewMuxer()
	)

	{{- range .services }}
	{
		if b, ok := enabled[{{ .ServerAlias }}.ServiceName]; len(enabled) == 0 || ok && b {
			{{ .ServerAlias }}Svc := {{ .NewServer }}(logger)
			{{ .ServerAlias }}Endpoints := {{ .ServerAlias }}.NewEndpoints({{ .ServerAlias }}Svc)
			{{ .ServerAlias }}Server := {{ .HttpServerAlias }}.New({{ .ServerAlias }}Endpoints, mux, dec, enc, eh, nil)
			{{ .HttpServerAlias }}.Mount(mux, {{ .ServerAlias }}Server)
		}
	}
	{{- end }}

	var handler http.Handler = mux
	{
		handler = httpmdlwr.Log(adapter)(handler)
		handler = httpmdlwr.RequestID()(handler)
	}
	return handler, mux
}

// errorHandler returns a function that writes and logs the given error.
// The function also writes and logs the error unique ID so that it's possible
// to correlate.
func errorHandler(logger *log.Logger) func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		id := ctx.Value(middleware.RequestIDKey).(string)
		_, _ = w.Write([]byte("[" + id + "] encoding: " + err.Error()))
		logger.Logf(log.ErrorLevel, "[%s] ERROR: %s", id, err.Error())
	}
}
`
