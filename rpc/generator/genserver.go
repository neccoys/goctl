package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	conf "github.com/neccoys/goctl/config"
	"github.com/neccoys/goctl/rpc/parser"
	"github.com/neccoys/goctl/util"
	"github.com/neccoys/goctl/util/format"
	"github.com/neccoys/goctl/util/pathx"
	"github.com/neccoys/goctl/util/stringx"
	"github.com/zeromicro/go-zero/core/collection"
)

const (
	serverTemplate = `{{.head}}

package server

import (
	{{if .notStream}}"context"{{end}}
 	{{if .check}}"google.golang.org/grpc/health/grpc_health_v1"{{end}}
	{{.imports}}
)

type {{.server}}Server struct {
	svcCtx *svc.ServiceContext
	{{.unimplementedServer}}
}

func New{{.server}}Server(svcCtx *svc.ServiceContext) *{{.server}}Server {
	return &{{.server}}Server{
		svcCtx: svcCtx,
	}
}

{{if .check}}
func (s *TxPayServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (s *TxPayServer) Watch(req *grpc_health_v1.HealthCheckRequest, w grpc_health_v1.Health_WatchServer) error {
	return nil
}
{{end}}

{{.funcs}}
`
	functionTemplate = `
{{if .hasComment}}{{.comment}}{{end}}
func (s *{{.server}}Server) {{.method}} ({{if .notStream}}ctx context.Context,{{if .hasReq}} in {{.request}}{{end}}{{else}}{{if .hasReq}} in {{.request}},{{end}}stream {{.streamBody}}{{end}}) ({{if .notStream}}{{.response}},{{end}}error) {
	l := logic.New{{.logicName}}({{if .notStream}}ctx,{{else}}stream.Context(),{{end}}s.svcCtx)
	return l.{{.method}}({{if .hasReq}}in{{if .stream}} ,stream{{end}}{{else}}{{if .stream}}stream{{end}}{{end}})
}
`
)

// GenServer generates rpc server file, which is an implementation of rpc server
func (g *DefaultGenerator) GenServer(ctx DirContext, proto parser.Proto, cfg *conf.Config, consul string) error {
	dir := ctx.GetServer()
	logicImport := fmt.Sprintf(`"%v"`, ctx.GetLogic().Package)
	svcImport := fmt.Sprintf(`"%v"`, ctx.GetSvc().Package)
	pbImport := fmt.Sprintf(`"%v"`, ctx.GetPb().Package)

	imports := collection.NewSet()
	imports.AddStr(logicImport, svcImport, pbImport)

	head := util.GetHead(proto.Name)
	service := proto.Service
	serverFilename, err := format.FileNamingFormat(cfg.NamingFormat, service.Name+"_server")
	if err != nil {
		return err
	}

	serverFile := filepath.Join(dir.Filename, serverFilename+".go")
	funcList, err := g.genFunctions(proto.PbPackage, service)
	if err != nil {
		return err
	}

	text, err := pathx.LoadTemplate(category, serverTemplateFile, serverTemplate)
	if err != nil {
		return err
	}

	notStream := false
	for _, rpc := range service.RPC {
		if !rpc.StreamsRequest && !rpc.StreamsReturns {
			notStream = true
			break
		}
	}

	var check string
	if consul == "grpc" {
		check = "grpc"
	}

	err = util.With("server").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
		"head":                head,
		"consul":              consul,
		"check":               check,
		"unimplementedServer": fmt.Sprintf("%s.Unimplemented%sServer", proto.PbPackage, stringx.From(service.Name).ToCamel()),
		"server":              stringx.From(service.Name).ToCamel(),
		"imports":             strings.Join(imports.KeysStr(), pathx.NL),
		"funcs":               strings.Join(funcList, pathx.NL),
		"notStream":           notStream,
	}, serverFile, true)
	return err
}

func (g *DefaultGenerator) genFunctions(goPackage string, service parser.Service) ([]string, error) {
	var functionList []string
	for _, rpc := range service.RPC {
		text, err := pathx.LoadTemplate(category, serverFuncTemplateFile, functionTemplate)
		if err != nil {
			return nil, err
		}

		comment := parser.GetComment(rpc.Doc())
		streamServer := fmt.Sprintf("%s.%s_%s%s", goPackage, parser.CamelCase(service.Name), parser.CamelCase(rpc.Name), "Server")
		buffer, err := util.With("func").Parse(text).Execute(map[string]interface{}{
			"server":     stringx.From(service.Name).ToCamel(),
			"logicName":  fmt.Sprintf("%sLogic", stringx.From(rpc.Name).ToCamel()),
			"method":     parser.CamelCase(rpc.Name),
			"request":    fmt.Sprintf("*%s.%s", goPackage, parser.CamelCase(rpc.RequestType)),
			"response":   fmt.Sprintf("*%s.%s", goPackage, parser.CamelCase(rpc.ReturnsType)),
			"hasComment": len(comment) > 0,
			"comment":    comment,
			"hasReq":     !rpc.StreamsRequest,
			"stream":     rpc.StreamsRequest || rpc.StreamsReturns,
			"notStream":  !rpc.StreamsRequest && !rpc.StreamsReturns,
			"streamBody": streamServer,
		})
		if err != nil {
			return nil, err
		}

		functionList = append(functionList, buffer.String())
	}
	return functionList, nil
}
