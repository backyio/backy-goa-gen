package logrus

import (
	"path"
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
	codegen.RegisterPluginFirst("logrus", "gen", nil, Generate)
	codegen.RegisterPluginLast("logrus-updater", "example", nil, UpdateExample)
}

// Generate generates logrus logger specific files.
func Generate(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
	for _, root := range roots {
		if r, ok := root.(*expr.RootExpr); ok {
			files = append(files, GenerateFiles(genpkg, r)...)
		}
	}
	return files, nil
}

// UpdateExample modifies the example generated files by replacing
// the log import reference when needed
// It also modify the initially generated main and service files
func UpdateExample(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {

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

// GenerateFiles create log specific files
func GenerateFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	fw := make([]*codegen.File, 1)
	fw[0] = GenerateLoggerFile(genpkg)
	return fw
}

// GenerateLoggerFile returns the generated logrus logger file.
func GenerateLoggerFile(genpkg string) *codegen.File {
	path := filepath.Join(codegen.Gendir, "log", "logger.go")
	title := "Logrus logger implementation"
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "log", []*codegen.ImportSpec{
			{Path: "github.com/sirupsen/logrus"},
			{Path: "fmt"},
		}),
	}

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "logruslogger",
		Source: loggerT,
	})

	return &codegen.File{Path: path, SectionTemplates: sections}
}

func updateExampleFile(genpkg string, root *expr.RootExpr, f *fileToModify) {

	header := f.file.SectionTemplates[0]
	logPath := path.Join(genpkg, "log")

	data := header.Data.(map[string]interface{})
	specs := data["Imports"].([]*codegen.ImportSpec)

	for _, spec := range specs {
		if spec.Path == "log" {
			spec.Name = "log"
			spec.Path = logPath
		}
	}

	if f.isMain {

		for _, s := range f.file.SectionTemplates {
			s.Source = strings.Replace(s.Source, `logger = log.New(os.Stderr, "[{{ .APIPkg }}] ", log.Ltime)`, `logger = logrus.New("[{{ .APIPkg }}]", true)`, 1)
			s.Source = strings.Replace(s.Source, "adapter = middleware.NewLogger(logger)", "adapter = logger", 1)
			s.Source = strings.Replace(s.Source, "handler = middleware.RequestID()(handler)",
				`handler = middleware.PopulateRequestContext()(handler)
				handler = middleware.RequestID(middleware.UseXRequestIDHeaderOption(true))(handler)`, 1)
			s.Source = strings.Replace(s.Source, `logger.Printf("[%s] ERROR: %s", id, err.Error())`,
				`logger.WithField("id",id).Error(err.Error())`, 1)
			s.Source = strings.Replace(s.Source, "logger.Print(", "logger.Info(", -1)
			s.Source = strings.Replace(s.Source, "logger.Printf(", "logger.Infof(", -1)
			s.Source = strings.Replace(s.Source, "logger.Println(", "logger.Info(", -1)
		}
	} else {
		for _, s := range f.file.SectionTemplates {
			s.Source = strings.Replace(s.Source, "logger.Print(", "logger.Info(", -1)
			s.Source = strings.Replace(s.Source, "logger.Printf(", "logger.Infof(", -1)
			s.Source = strings.Replace(s.Source, "logger.Println(", "logger.Info(", -1)
		}
	}
}

const loggerT = `
// Logger is an adapted logrus logger
type Logger struct {
	*logrus.Logger
    Service string
}

// New creates a new logrus logger
func New(serviceName string, isDebug bool) *Logger {
	l := logrus.New()
    l.SetLevel(logrus.ErrorLevel)
	if isDebug {
		l.SetLevel(logrus.FatalLevel)
	}

	return &Logger{
		Logger: l,
		Service: serviceName,
	}
}

// Log is called by the log middleware to log HTTP requests key values
func (logger *Logger) Log(keyvals ...interface{}) error {
	fields := FormatFields(keyvals)
	logger.WithFields(fields).Info("HTTP Request")
	return nil
}

// formatFields formats input keyvals
// ref: https://github.com/goadesign/goa/blob/v1/logging/logrus/adapter.go#L64
func FormatFields(keyvals []interface{}) logrus.Fields {
	n := (len(keyvals) + 1) / 2
	res := make(logrus.Fields, n)
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		var v interface{} 
		if i+1 < len(keyvals) {
			v = keyvals[i+1]
		}
		res[fmt.Sprintf("%v", k)] = v
	}
	return res
}
`
