package generator

import (
	conf "github.com/neccoys/goctl/config"
	"github.com/neccoys/goctl/rpc/parser"
)

// Generator defines a generator interface to describe how to generate rpc service
type Generator interface {
	Prepare() error
	GenMain(ctx DirContext, proto parser.Proto, cfg *conf.Config, consul string) error
	GenCall(ctx DirContext, proto parser.Proto, cfg *conf.Config) error
	GenEtc(ctx DirContext, proto parser.Proto, cfg *conf.Config, consul string) error
	GenConfig(ctx DirContext, proto parser.Proto, cfg *conf.Config, consul string) error
	GenLogic(ctx DirContext, proto parser.Proto, cfg *conf.Config) error
	GenServer(ctx DirContext, proto parser.Proto, cfg *conf.Config, consul string) error
	GenSvc(ctx DirContext, proto parser.Proto, cfg *conf.Config) error
	GenPb(ctx DirContext, protoImportPath []string, proto parser.Proto, cfg *conf.Config, c *ZRpcContext, goOptions ...string) error
}
