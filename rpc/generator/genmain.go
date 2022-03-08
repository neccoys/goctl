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
)

const mainTemplate = `package main

import (
	"flag"
	"fmt"
 	{{if .consul}}"github.com/neccoys/go-zero-extension/consul"{{end}}
    {{if .check}}"google.golang.org/grpc/health/grpc_health_v1"{{end}}
	"log"

	{{.imports}}

	"github.com/joho/godotenv"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
    configFile = flag.String("f", "etc/{{.serviceName}}.yaml", "the config file")
    envFile    = flag.String("env", "etc/.env", "the env file")
)

func main() {
	flag.Parse()
	if err := godotenv.Load(*envFile); err != nil {
		log.Fatal("Error loading .env file")
	}

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	ctx := svc.NewServiceContext(c)
	srv := server.New{{.serviceNew}}Server(ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		{{.pkg}}.Register{{.service}}Server(grpcServer, srv)
		{{if .check}}grpc_health_v1.RegisterHealthServer(grpcServer, srv){{end}}

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	{{if .consul}}
	// 注册Consul服务
    if err := consul.RegisterService(c.ListenOn, c.Consul); err != nil {
        log.Println("Consul Error:", err)
    }
    {{end}}

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
`

// GenMain generates the main file of the rpc service, which is an rpc service program call entry
func (g *DefaultGenerator) GenMain(ctx DirContext, proto parser.Proto, cfg *conf.Config, consul string) error {
	mainFilename, err := format.FileNamingFormat(cfg.NamingFormat, ctx.GetServiceName().Source())
	if err != nil {
		return err
	}

	fileName := filepath.Join(ctx.GetMain().Filename, fmt.Sprintf("%v.go", mainFilename))
	imports := make([]string, 0)
	pbImport := fmt.Sprintf(`"%v"`, ctx.GetPb().Package)
	svcImport := fmt.Sprintf(`"%v"`, ctx.GetSvc().Package)
	remoteImport := fmt.Sprintf(`"%v"`, ctx.GetServer().Package)
	configImport := fmt.Sprintf(`"%v"`, ctx.GetConfig().Package)
	imports = append(imports, configImport, pbImport, remoteImport, svcImport)
	text, err := pathx.LoadTemplate(category, mainTemplateFile, mainTemplate)
	if err != nil {
		return err
	}

	etcFileName, err := format.FileNamingFormat(cfg.NamingFormat, ctx.GetServiceName().Source())
	if err != nil {
		return err
	}

	var check string
	if consul == "grpc" {
		check = "grpc"
	}

	return util.With("main").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
		"serviceName": etcFileName,
		"imports":     strings.Join(imports, pathx.NL),
		"consul":      consul,
		"check":       check,
		"pkg":         proto.PbPackage,
		"serviceNew":  stringx.From(proto.Service.Name).ToCamel(),
		"service":     parser.CamelCase(proto.Service.Name),
	}, fileName, false)
}
