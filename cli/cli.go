package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/blakewilliams/customs"
	"github.com/blakewilliams/customs/formatters/prettyformat"
	"github.com/blakewilliams/customs/inspectors"
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
				Name:  "inspect",
				Usage: "todo",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "Uses provided config `FILE`",
					},
					&cli.StringFlag{
						Name:    "diffSource",
						Aliases: []string{"d"},
						Usage:   "Uses the provided diff `FILE`",
					},
					&cli.BoolFlag{
						Name:  "only-import-json",
						Usage: "Outputs only the JSON and does not run the inspectors",
					},
				},
				Action: func(cctx *cli.Context) error {
					var in io.Reader

					fi, err := os.Stdin.Stat()
					if err != nil {
						panic(err)
					}
					if (fi.Mode() & os.ModeCharDevice) == 0 {
						in = os.Stdin
					} else if cctx.String("diffSource") != "" {
						f, err := os.Open(cctx.String("diffSource"))
						if err != nil {
							return err
						}
						defer f.Close()
						in = f
					} else {
						return cli.ShowCommandHelp(cctx, "run")
					}

					config := customs.Configuration{
						Formatter: prettyformat.New(os.Stdout),
					}

					inspection, err := customs.NewInspection(config, in)
					if err != nil {
						fmt.Println(err)
						return cli.ShowCommandHelp(cctx, "run")
					}

					if cctx.Bool("only-import-json") {
						out, err := inspection.ImportJSON()
						if err != nil {
							fmt.Printf("Could not return import JSON: %s\n", err)
						}

						fmt.Println(string(out))
						return nil
					}

					inspection.Perform()

					return nil
				},
			},
			{
				Name:  "inspector",
				Usage: "runs the given built-in inspector",
				Subcommands: []*cli.Command{
					{
						Name:  "rails_job_perform",
						Usage: "Runs the Rails job inspector to ensure perform is modified safely for rolling deploys",
						Action: func(cctx *cli.Context) error {
							err := inspectors.Wrap("rails_job_perform", inspectors.RailsJobArguments)
							if err != nil {
								// TODO write {} with error to stdout
								fmt.Fprintf(os.Stderr, "%s\n", err)
							}
							return nil
						},
					},
				},
			},
		},
	}

	return &CLI{app: app}
}

func (c *CLI) Run(args []string) error {
	return c.app.Run(args)
}
