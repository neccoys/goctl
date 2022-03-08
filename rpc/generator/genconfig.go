package generator

import (
	conf "github.com/neccoys/goctl/config"
	"github.com/neccoys/goctl/rpc/parser"
	"github.com/neccoys/goctl/util"
	"github.com/neccoys/goctl/util/format"
	"github.com/neccoys/goctl/util/pathx"
	"path/filepath"
)

const configTemplate = `package config

import (
	"github.com/zeromicro/go-zero/zrpc"
	{{if .consul}}"github.com/neccoys/go-zero-extension/consul"{{end}}
)

type Config struct {
	zrpc.RpcServerConf
	{{if .consul}}Consul consul.Conf{{end}}
}
`

// GenConfig generates the configuration structure definition file of the rpc service,
// which contains the zrpc.RpcServerConf configuration item by default.
// You can specify the naming style of the target file name through config.Config. For details,
// see https://github.com/zeromicro/go-zero/tree/master/tools/goctl/config/config.go
func (g *DefaultGenerator) GenConfig(ctx DirContext, _ parser.Proto, cfg *conf.Config, consul string) error {
	dir := ctx.GetConfig()
	configFilename, err := format.FileNamingFormat(cfg.NamingFormat, "config")
	if err != nil {
		return err
	}

	fileName := filepath.Join(dir.Filename, configFilename+".go")
	if pathx.FileExists(fileName) {
		return nil
	}

	text, err := pathx.LoadTemplate(category, configTemplateFileFile, configTemplate)
	if err != nil {
		return err
	}

	return util.With(fileName).GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
		"consul": consul,
	}, fileName, false)
}
