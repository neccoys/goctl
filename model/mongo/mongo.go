package mongo

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
	"github.com/neccoys/goctl/config"
	"github.com/neccoys/goctl/model/mongo/generate"
	file "github.com/neccoys/goctl/util"
	"github.com/neccoys/goctl/util/pathx"
)

// Action provides the entry for goctl mongo code generation.
func Action(ctx *cli.Context) error {
	tp := ctx.StringSlice("type")
	c := ctx.Bool("cache")
	o := strings.TrimSpace(ctx.String("dir"))
	s := ctx.String("style")
	home := ctx.String("home")
	remote := ctx.String("remote")
	if len(remote) > 0 {
		repo, _ := file.CloneIntoGitHome(remote)
		if len(repo) > 0 {
			home = repo
		}
	}
	if len(home) > 0 {
		pathx.RegisterGoctlHome(home)
	}

	if len(tp) == 0 {
		return errors.New("missing type")
	}

	cfg, err := config.NewConfig(s)
	if err != nil {
		return err
	}

	a, err := filepath.Abs(o)
	if err != nil {
		return err
	}

	return generate.Do(&generate.Context{
		Types:  tp,
		Cache:  c,
		Output: a,
		Cfg:    cfg,
	})
}
