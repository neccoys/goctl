package gogen

import (
	"fmt"
	"path"
	"strings"

	"github.com/neccoys/goctl/api/spec"
	"github.com/neccoys/goctl/config"
	"github.com/neccoys/goctl/internal/version"
	"github.com/neccoys/goctl/util"
	"github.com/neccoys/goctl/util/format"
	"github.com/neccoys/goctl/util/pathx"
	"github.com/neccoys/goctl/vars"
)

const (
	defaultLogicPackage = "logic"
	handlerTemplate     = `package {{.PkgName}}

import (
	"net/http"
	{{if .HasRequest}}"{{.CommonPath}}/common/vaildx"
    {{end}}"{{.CommonPath}}/common/responsex"
    {{if .HasRequest}}"encoding/json"
	{{end}}{{if .After1_1_10}}{{if .HasRequest}}"github.com/zeromicro/go-zero/rest/httpx"{{end}}{{end}}
    {{if .HasRequest}}"go.opentelemetry.io/otel/attribute"
    {{end}}"go.opentelemetry.io/otel/trace"
	{{.ImportPackages}}
)

func {{.HandlerName}}(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		span := trace.SpanFromContext(r.Context())
        defer span.End()

		{{if .HasRequest}}var req types.{{.RequestType}}

        if err := httpx.ParseJsonBody(r, &req); err != nil {
            responsex.Json(w, r, responsex.FAIL, nil, err)
            return
        }

		if err := vaildx.Validator.Struct(req); err != nil {
			responsex.Json(w, r, responsex.INVALID_PARAMETER, nil, err)
			return
		}

		if requestBytes, err := json.Marshal(req); err == nil {
            span.SetAttributes(attribute.KeyValue{
                Key:   "request",
                Value: attribute.StringValue(string(requestBytes)),
            })
        }

		{{end}}l := {{.LogicName}}.New{{.LogicType}}(r.Context(), ctx)
		{{if .HasResp}}resp, {{end}}err := l.{{.Call}}({{if .HasRequest}}&req{{end}})
		if err != nil {
			responsex.Json(w, r, err.Error(), nil, err)
		} else {
			{{if .HasResp}}responsex.Json(w, r, responsex.SUCCESS, resp, err){{else}}responsex.Json(w, r, responsex.SUCCESS, nil, err){{end}}
		}
	}
}

`
)

type handlerInfo struct {
	CommonPath     string
	PkgName        string
	ImportPackages string
	HandlerName    string
	RequestType    string
	LogicName      string
	LogicType      string
	Call           string
	HasResp        bool
	HasRequest     bool
	After1_1_10    bool
}

func genHandler(dir, rootPkg string, cfg *config.Config, group spec.Group, route spec.Route) error {
	handler := getHandlerName(route)
	handlerPath := getHandlerFolderPath(group, route)
	pkgName := handlerPath[strings.LastIndex(handlerPath, "/")+1:]
	logicName := defaultLogicPackage
	if handlerPath != handlerDir {
		handler = strings.Title(handler)
		logicName = pkgName
	}
	parentPkg, err := getParentPackage(dir)
	if err != nil {
		return err
	}

	goctlVersion := version.GetGoctlVersion()
	// todo(anqiansong): This will be removed after a certain number of production versions of goctl (probably 5)
	after1_1_10 := version.IsVersionGreaterThan(goctlVersion, "1.1.10")
	return doGenToFile(dir, handler, cfg, group, route, handlerInfo{
		CommonPath:     rootPkg,
		PkgName:        pkgName,
		ImportPackages: genHandlerImports(group, route, parentPkg),
		HandlerName:    handler,
		After1_1_10:    after1_1_10,
		RequestType:    util.Title(route.RequestTypeName()),
		LogicName:      logicName,
		LogicType:      strings.Title(getLogicName(route)),
		Call:           strings.Title(strings.TrimSuffix(handler, "Handler")),
		HasResp:        len(route.ResponseTypeName()) > 0,
		HasRequest:     len(route.RequestTypeName()) > 0,
	})
}

func doGenToFile(dir, handler string, cfg *config.Config, group spec.Group,
	route spec.Route, handleObj handlerInfo) error {
	filename, err := format.FileNamingFormat(cfg.NamingFormat, handler)
	if err != nil {
		return err
	}

	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          getHandlerFolderPath(group, route),
		filename:        filename + ".go",
		templateName:    "handlerTemplate",
		category:        category,
		templateFile:    handlerTemplateFile,
		builtinTemplate: handlerTemplate,
		data:            handleObj,
	})
}

func genHandlers(dir, rootPkg string, cfg *config.Config, api *spec.ApiSpec) error {
	for _, group := range api.Service.Groups {
		for _, route := range group.Routes {
			if err := genHandler(dir, rootPkg, cfg, group, route); err != nil {
				return err
			}
		}
	}

	return nil
}

func genHandlerImports(group spec.Group, route spec.Route, parentPkg string) string {
	var imports []string
	imports = append(imports, fmt.Sprintf("\"%s\"",
		pathx.JoinPackages(parentPkg, getLogicFolderPath(group, route))))
	imports = append(imports, fmt.Sprintf("\"%s\"", pathx.JoinPackages(parentPkg, contextDir)))
	if len(route.RequestTypeName()) > 0 {
		imports = append(imports, fmt.Sprintf("\"%s\"\n", pathx.JoinPackages(parentPkg, typesDir)))
	}

	currentVersion := version.GetGoctlVersion()
	// todo(anqiansong): This will be removed after a certain number of production versions of goctl (probably 5)
	if !version.IsVersionGreaterThan(currentVersion, "1.1.10") {
		imports = append(imports, fmt.Sprintf("\"%s/rest/httpx\"", vars.ProjectOpenSourceURL))
	}

	return strings.Join(imports, "\n\t")
}

func getHandlerBaseName(route spec.Route) (string, error) {
	handler := route.Handler
	handler = strings.TrimSpace(handler)
	handler = strings.TrimSuffix(handler, "handler")
	handler = strings.TrimSuffix(handler, "Handler")
	return handler, nil
}

func getHandlerFolderPath(group spec.Group, route spec.Route) string {
	folder := route.GetAnnotation(groupProperty)
	if len(folder) == 0 {
		folder = group.GetAnnotation(groupProperty)
		if len(folder) == 0 {
			return handlerDir
		}
	}

	folder = strings.TrimPrefix(folder, "/")
	folder = strings.TrimSuffix(folder, "/")
	return path.Join(handlerDir, folder)
}

func getHandlerName(route spec.Route) string {
	handler, err := getHandlerBaseName(route)
	if err != nil {
		panic(err)
	}

	return handler + "Handler"
}

func getLogicName(route spec.Route) string {
	handler, err := getHandlerBaseName(route)
	if err != nil {
		panic(err)
	}

	return handler + "Logic"
}
