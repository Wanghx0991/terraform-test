package biz

import (
	"encoding/json"
	"github.com/aliyun/terraform-test/biz/module"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

func init() {
	customFormatter := new(log.TextFormatter)
	customFormatter.FullTimestamp = true
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.DisableTimestamp = false
	customFormatter.DisableColors = false
	customFormatter.ForceColors = true
	log.SetFormatter(customFormatter)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
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
					&cli.StringFlag{Name: "stage", Aliases: []string{"s"}, Required: false, Usage: "The Stage of terraform compatibility. valid value: [PrevStage,NextStage,NewVersion]"},
				},
				Action: func(c *cli.Context) error {
					raw := strings.TrimSpace(c.String("relative_resources"))
					log.Infof("The relative resources: %s", raw)
					resources := make([]string, 0)
					if len(raw) != 0 {
						resources = strings.Split(raw, ",")
					}
					mod, err := module.ExecuteModules(resources, c.String("stage"))
					if err != nil {
						log.Error(err)
						return err
					}
					res, err := json.Marshal(mod)
					if err != nil {
						log.Error(err)
						return err
					}
					log.Infof("%v", string(res))
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
