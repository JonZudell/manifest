package cli

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

type CLI struct {
	app *cli.App
}

func New() *CLI {
	app := &cli.App{
		Name:  "customs",
		Usage: "Runs rules against pull requests and diffs",
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "todo",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "Uses provided config `FILE`",
					},
				},
				Action: func(cctx *cli.Context) error {
					fmt.Println("TODO")
					return nil
				},
			},
		},
	}

	return &CLI{app: app}
}

func (c *CLI) Run(args []string) error {
	return c.app.Run(args)
}
