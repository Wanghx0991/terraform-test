package biz

import (
	"encoding/json"
	"github.com/aliyun/terraform-test/biz/module"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strings"
)

func init() {
	logrus.SetReportCaller(true)
}

func Run() {
	app := &cli.App{
		Name:        "Relative Module Test && Code Coverage",
		Usage:       " A CLI of terraform test for testing relative modules and code coverage",
		Description: "A CLI of terraform test for testing relative modules and code coverage",

		Commands: []*cli.Command{
			{
				Name:  "module_test",
				Usage: "Relative Module Test",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "relative_resources", Aliases: []string{"r"}, Required: true, Usage: "Input relative_resources"},
					&cli.StringFlag{Name: "terraform_parameter", Aliases: []string{"p"}, Required: false, Usage: "terraform module's parameter"},
				},
				Action: func(c *cli.Context) error {
					raw := strings.TrimSpace(c.String("relative_resources"))
					logrus.Infof("The relative resources: %s", raw)
					resources := make([]string, 0)
					if len(raw) != 0 {
						resources = strings.Split(raw, ",")
					}
					para := c.String("terraform_parameter")
					mod, err := module.ExecuteModules(resources, para)
					if err != nil {
						logrus.Error(err)
						return err
					}
					res, err := json.Marshal(mod)
					if err != nil {
						logrus.Error(err)
						return err
					}
					logrus.Infof("%v", string(res))
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
