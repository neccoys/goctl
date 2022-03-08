package gogen

import (
	"fmt"
	"github.com/neccoys/goctl/api/spec"
	"github.com/neccoys/goctl/config"
	"github.com/neccoys/goctl/util/format"
	"strings"
)

const makefileTemplate = `t := {{.pkgName}}
lang:
	easyi18n generate --pkg=locales ./locales ./locales/locales.go

api:
	goctl api go -api $(t)/$(t).api -dir $(t) -remote https://github.com/neccohuang/go-zero-template

run:
	go run $(t)/$(t).go -f $(t)/etc/$(t).yaml -env $(t)/etc/.env
`

func genMakefile(rootPkg string, cfg *config.Config, api *spec.ApiSpec, params map[string]interface{}) error {

	name := strings.ToLower(api.Service.Name)
	pkgName, err := format.FileNamingFormat(cfg.NamingFormat, name)
	if err != nil {
		return err
	}

	configName := pkgName
	if strings.HasSuffix(pkgName, "-api") {
		pkgName = strings.ReplaceAll(pkgName, "-api", "")
	}

	var commonPath string
	if _, ok := params["commonPath"]; !ok || params["commonPath"] == "" {
		commonPath = "../"
	} else {
		commonPath = fmt.Sprintf("%s", params["commonPath"])
	}

	return genFile(fileGenConfig{
		dir:             commonPath,
		subdir:          "/",
		filename:        "Makefile",
		templateName:    "stateTemplate",
		category:        category,
		templateFile:    "makefile.tpl",
		builtinTemplate: makefileTemplate,
		data: map[string]string{
			"pkgName":    pkgName,
			"configName": configName,
		},
	})
}
