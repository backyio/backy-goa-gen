package logrus

import (
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type fileToModify struct {
	file        *codegen.File
	path        string
	serviceName string
	isMain      bool
}

// Register the plugin Generator functions.
func init() {
	codegen.RegisterPluginLast("micro-log", "example", nil, UpdateExample)
}

var (
	middlewarePath string
)

// getMiddlewarePath generate path for middleware from pkg name
func getMiddlewarePath(genpkg string) string {
	p := strings.Replace(genpkg, "\\", "/", -1)
	p = strings.Replace(p, "endpoint/", "", -1)
	p = filepath.Join(p, "middleware")
	return strings.Replace(p, "\\", "/", -1)
}

// UpdateExample modifies the example generated files by replacing
// the log import reference when needed
// It also modify the initially generated main and service files
func UpdateExample(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
	middlewarePath = getMiddlewarePath(genpkg)

	filesToModify := []*fileToModify{}

	for _, root := range roots {
		if r, ok := root.(*expr.RootExpr); ok {

			// Add the generated main files
			for _, svr := range r.API.Servers {
				pkg := codegen.SnakeCase(codegen.Goify(svr.Name, true))
				filesToModify = append(filesToModify,
					&fileToModify{path: filepath.Join("cmd", pkg, "main.go"), serviceName: svr.Name, isMain: true})
				filesToModify = append(filesToModify,
					&fileToModify{path: filepath.Join("cmd", pkg, "http.go"), serviceName: svr.Name, isMain: true})
				filesToModify = append(filesToModify,
					&fileToModify{path: filepath.Join("cmd", pkg, "grpc.go"), serviceName: svr.Name, isMain: true})
			}

			// Add the generated service files
			for _, svc := range r.API.HTTP.Services {
				servicePath := codegen.SnakeCase(svc.Name()) + ".go"
				filesToModify = append(filesToModify, &fileToModify{path: servicePath, serviceName: svc.Name(), isMain: false})
			}

			// Update the added files
			for _, fileToModify := range filesToModify {
				for _, file := range files {
					if file.Path == fileToModify.path {
						fileToModify.file = file
						updateExampleFile(genpkg, r, fileToModify)
						break
					}
				}
			}
		}
	}
	return files, nil
}

func updateExampleFile(genpkg string, root *expr.RootExpr, f *fileToModify) {

	header := f.file.SectionTemplates[0]

	data := header.Data.(map[string]interface{})
	specs := data["Imports"].([]*codegen.ImportSpec)

	for _, spec := range specs {
		if spec.Path == "log" {
			spec.Name = "mlog"
			spec.Path = "go-micro.dev/v4/logger"
		}
		if spec.Name == "httpmdlwr" {
			spec.Path = middlewarePath
		}
	}

	if f.isMain {
		for _, s := range f.file.SectionTemplates {
			s.Source = strings.Replace(s.Source, `logger = log.New(os.Stderr, "[{{ .APIPkg }}] ", log.Ltime)`, "", 1)
			s.Source = strings.Replace(s.Source, `logger *log.Logger,`, "logger mlog.Logger,", 1)
			s.Source = strings.Replace(s.Source, `errorHandler(logger *log.Logger) func`, "errorHandler(logger mlog.Logger) func", 1)
			s.Source = strings.Replace(s.Source, "adapter middleware.Logger", "adapter mlog.Logger", 1)
			s.Source = strings.Replace(s.Source, "adapter = middleware.NewLogger(logger)", "adapter = logger", 1)
			s.Source = strings.Replace(s.Source, "id := ctx.Value(middleware.RequestIDKey).(string)", "id := ctx.Value(httpmdlwr.RequestIDKey).(string)", 1)

			s.Source = strings.Replace(s.Source, `// errorHandler returns a function that writes and logs the given error.
// The function also writes and logs the error unique ID so that it's possible
// to correlate.
func errorHandler(logger mlog.Logger) func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		id := ctx.Value(httpmdlwr.RequestIDKey).(string)
		_, _ = w.Write([]byte("[" + id + "] encoding: " + err.Error()))
		logger.Logf(mlog.ErrorLevel, "[%s] ERROR: %s", id, err.Error())
	}
}`, ``, 1)
			s.Source = strings.Replace(s.Source, "logger.Print(", "logger.Log(mlog.InfoLevel,", -1)
			s.Source = strings.Replace(s.Source, "logger.Printf(", "logger.Log(mlog.InfoLevel,", -1)
			s.Source = strings.Replace(s.Source, "logger.Println(", "logger.Log(mlog.InfoLevel,", -1)
			s.Source = strings.Replace(s.Source, "eh := errorHandler(logger)", "eh := httpmdlwr.ErrorHandler(logger)", -1)
		}
	} else {
		for _, s := range f.file.SectionTemplates {
			s.Source = strings.Replace(s.Source, `logger *log.Logger`, "logger mlog.Logger", 1)
			s.Source = strings.Replace(s.Source, "logger.Print(", "logger.Log(mlog.InfoLevel,", -1)
			s.Source = strings.Replace(s.Source, "logger.Printf(", "logger.Log(mlog.InfoLevel", -1)
			s.Source = strings.Replace(s.Source, "logger.Println(", "logger.Log(mlog.InfoLevel,", -1)
		}
	}
}
