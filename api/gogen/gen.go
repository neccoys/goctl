package gogen

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/logrusorgru/aurora"
	apiformat "github.com/neccoys/goctl/api/format"
	"github.com/neccoys/goctl/api/parser"
	apiutil "github.com/neccoys/goctl/api/util"
	"github.com/neccoys/goctl/config"
	"github.com/neccoys/goctl/util"
	"github.com/neccoys/goctl/util/pathx"
	"github.com/urfave/cli"
	"github.com/zeromicro/go-zero/core/logx"
)

const tmpFile = "%s-%d"

var tmpDir = path.Join(os.TempDir(), "goctl")

// GoCommand gen go project files from command line
func GoCommand(c *cli.Context) error {
	apiFile := c.String("api")
	dir := c.String("dir")
	namingStyle := c.String("style")
	home := c.String("home")
	remote := c.String("remote")
	if len(remote) > 0 {
		repo, _ := util.CloneIntoGitHome(remote)
		if len(repo) > 0 {
			home = repo
		}
	}

	if len(home) > 0 {
		pathx.RegisterGoctlHome(home)
	}
	if len(apiFile) == 0 {
		return errors.New("missing -api")
	}
	if len(dir) == 0 {
		return errors.New("missing -dir")
	}

	params := make(map[string]interface{})

	if c.String("common") != "" {
		params["commonPath"] = c.String("common")
	}

	if len(params) > 0 {
		return DoGenProject(apiFile, dir, namingStyle, params)
	}

	return DoGenProject(apiFile, dir, namingStyle)
}

// DoGenProject gen go project files with api file
func DoGenProject(apiFile, dir, style string, i ...interface{}) error {
	api, err := parser.Parse(apiFile)
	if err != nil {
		return err
	}

	cfg, err := config.NewConfig(style)
	if err != nil {
		return err
	}

	logx.Must(pathx.MkdirIfNotExist(dir))
	rootPkg, err := getParentPackage(dir)
	if err != nil {
		return err
	}

	modeName := strings.Split(rootPkg, "/"+api.Service.Name)

	logx.Must(genEtc(dir, cfg, api))
	logx.Must(genConfig(dir, cfg, api))
	logx.Must(genMain(dir, rootPkg, cfg, api))
	logx.Must(genServiceContext(dir, rootPkg, cfg, api))
	logx.Must(genTypes(dir, cfg, api))
	logx.Must(genRoutes(dir, rootPkg, cfg, api))
	logx.Must(genHandlers(dir, modeName[0], cfg, api))
	logx.Must(genLogic(dir, rootPkg, cfg, api))
	logx.Must(genMiddleware(dir, cfg, api))

	if len(i) > 0 {
		logx.Must(genErrorx(rootPkg, i[0].(map[string]interface{})))
		logx.Must(genVaildx(rootPkg, i[0].(map[string]interface{})))
		logx.Must(genState(rootPkg, i[0].(map[string]interface{})))
		logx.Must(genResponse(modeName[0], i[0].(map[string]interface{})))
		logx.Must(genMakefile(rootPkg, cfg, api, i[0].(map[string]interface{})))
	}

	if err := backupAndSweep(apiFile); err != nil {
		return err
	}

	if err := apiformat.ApiFormatByPath(apiFile); err != nil {
		return err
	}

	fmt.Println(aurora.Green("Done."))
	return nil
}

func backupAndSweep(apiFile string) error {
	var err error
	var wg sync.WaitGroup

	wg.Add(2)
	_ = os.MkdirAll(tmpDir, os.ModePerm)

	go func() {
		_, fileName := filepath.Split(apiFile)
		_, e := apiutil.Copy(apiFile, fmt.Sprintf(path.Join(tmpDir, tmpFile), fileName, time.Now().Unix()))
		if e != nil {
			err = e
		}
		wg.Done()
	}()
	go func() {
		if e := sweep(); e != nil {
			err = e
		}
		wg.Done()
	}()
	wg.Wait()

	return err
}

func sweep() error {
	keepTime := time.Now().AddDate(0, 0, -7)
	return filepath.Walk(tmpDir, func(fpath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		pos := strings.LastIndexByte(info.Name(), '-')
		if pos > 0 {
			timestamp := info.Name()[pos+1:]
			seconds, err := strconv.ParseInt(timestamp, 10, 64)
			if err != nil {
				// print error and ignore
				fmt.Println(aurora.Red(fmt.Sprintf("sweep ignored file: %s", fpath)))
				return nil
			}

			tm := time.Unix(seconds, 0)
			if tm.Before(keepTime) {
				if err := os.Remove(fpath); err != nil {
					fmt.Println(aurora.Red(fmt.Sprintf("failed to remove file: %s", fpath)))
					return err
				}
			}
		}

		return nil
	})
}
