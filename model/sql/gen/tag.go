package gen

import (
	"github.com/neccoys/goctl/model/sql/template"
	"github.com/neccoys/goctl/util"
	"github.com/neccoys/goctl/util/pathx"
)

func genTag(table Table, in string) (string, error) {
	if in == "" {
		return in, nil
	}

	text, err := pathx.LoadTemplate(category, tagTemplateFile, template.Tag)
	if err != nil {
		return "", err
	}

	output, err := util.With("tag").Parse(text).Execute(map[string]interface{}{
		"field": in,
		"data":  table,
	})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
