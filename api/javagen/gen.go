package javagen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/urfave/cli"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/neccoys/goctl/api/parser"
	"github.com/neccoys/goctl/util/pathx"
)

// JavaCommand the generate java code command entrance
func JavaCommand(c *cli.Context) error {
	apiFile := c.String("api")
	dir := c.String("dir")
	if len(apiFile) == 0 {
		return errors.New("missing -api")
	}
	if len(dir) == 0 {
		return errors.New("missing -dir")
	}

	api, err := parser.Parse(apiFile)
	if err != nil {
		return err
	}

	api.Service = api.Service.JoinPrefix()
	packetName := strings.TrimSuffix(api.Service.Name, "-api")
	logx.Must(pathx.MkdirIfNotExist(dir))
	logx.Must(genPacket(dir, packetName, api))
	logx.Must(genComponents(dir, packetName, api))

	fmt.Println(aurora.Green("Done."))
	return nil
}
