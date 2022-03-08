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

const etcTemplate = `Name: {{.serviceName}}.rpc
ListenOn: 127.0.0.1:8080
{{if .consul}}
Consul:
  Host: 127.0.0.1:8500
  Key: {{.serviceName}}.rpc
  Check: {{if .check}}grpc{{else}}ttl{{end}}
  Meta:
    Protocol: grpc
  Tag:
    - {{.serviceName}}
{{else}}
Etcd:
  Hosts:
  - 127.0.0.1:2379
  Key: {{.serviceName}}.rpc
{{end}}
`

// GenEtc generates the yaml configuration file of the rpc service,
// including host, port monitoring configuration items and etcd configuration
func (g *DefaultGenerator) GenEtc(ctx DirContext, _ parser.Proto, cfg *conf.Config, consul string) error {
	dir := ctx.GetEtc()
	etcFilename, err := format.FileNamingFormat(cfg.NamingFormat, ctx.GetServiceName().Source())
	if err != nil {
		return err
	}

	fileName := filepath.Join(dir.Filename, fmt.Sprintf("%v.yaml", etcFilename))

	text, err := pathx.LoadTemplate(category, etcTemplateFileFile, etcTemplate)
	if err != nil {
		return err
	}

	var check string
	if consul == "grpc" {
		check = "grpc"
	}

	return util.With("etc").Parse(text).SaveTo(map[string]interface{}{
		"serviceName": strings.ToLower(stringx.From(ctx.GetServiceName().Source()).ToCamel()),
		"consul":      consul,
		"check":       check,
	}, fileName, false)
}
