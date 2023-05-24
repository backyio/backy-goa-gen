// micromuxer is generate custom service handler for Go-Micro
package micromuxer

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type serversToModify struct {
	file        *codegen.File
	path        string
	serviceName string
	isMain      bool
}

// Register the plugin Generator functions.
func init() {
	codegen.RegisterPluginFirst("micro-muxer", "gen", nil, Generate)
}

// Generate generates go-muxer specific file.
func Generate(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
	var servers serversToModify
	for _, root := range roots {
		if r, ok := root.(*expr.RootExpr); ok {
			CollectServices(genpkg, r, &servers)
		}
	}
	files = append(files, GenerateMicroMuxerFile(genpkg, servers))
	return files, nil
}

// CollectServices collecting information about all services
func CollectServices(genpkg string, root eval.Root, servers *serversToModify) {
	if r, ok := root.(*expr.RootExpr); ok {

		// Add the generated main files
		for _, svr := range r.API.Servers {
			pkg := codegen.SnakeCase(codegen.Goify(svr.Name, true))
			fmt.Printf("package : %s \n", pkg)
		}
	}
}

// GenerateMicroMuxerFile returns the generated go muxer file.
func GenerateMicroMuxerFile(genpkg string, servers *serversToModify) *codegen.File {
	path := filepath.Join(codegen.Gendir, "micro.go")
	title := "Go-Micro muxer generator"
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "service", []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "net/http"},
			{Path: "go-micro.dev/v4/logger", Name: "mlog"},
			{Path: "goa.design/goa/v3/middleware"},
			{Path: "goa.design/goa/v3/http/middleware", "httpmdlwr"},
			{Path: "goa.design/goa/v3/http", "goahttp"},
		}),
	}

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "go-micro-muxer",
		Data:   servers,
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


	{
		if b, ok := enabled[discovery.ServiceName]; len(enabled) == 0 || ok && b {
			discoverySvc := NewDiscovery(logger)
			discoveryEndpoints := discovery.NewEndpoints(discoverySvc)
			discoveryServer := discoverysvr.New(discoveryEndpoints, mux, dec, enc, eh, nil)
			discoverysvr.Mount(mux, discoveryServer)
		}
	}

	var handler http.Handler = mux
	{
		handler = httpmdlwr.Log(adapter)(handler)
		handler = httpmdlwr.RequestID()(handler)
	}
	return handler, mux
}

// Log is called by the log middleware to log HTTP requests key values
func (logger *Logger) Log(keyvals ...interface{}) error {
	fields := FormatFields(keyvals)
	logger.Fields(fields).Log(mlog.InfoLevel, "HTTP Request")
	return nil
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
