package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/blakewilliams/customs"
	"github.com/blakewilliams/customs/formatters/prettyformat"
	"github.com/blakewilliams/customs/githelpers"
	"github.com/blakewilliams/customs/github"
	"github.com/blakewilliams/customs/inspectors"
	"github.com/fatih/color"
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
				Usage: "Runs the configured inspectors against the provided diff",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "Uses provided config `FILE`",
					},
					&cli.StringFlag{
						Name:    "diff",
						Aliases: []string{"d"},
						Usage:   "Uses the provided diff `FILE`",
					},
					&cli.BoolFlag{
						Name:  "only-import-json",
						Usage: "Outputs only the JSON and does not run the inspectors",
					},
					&cli.BoolFlag{
						Name:    "include-pr-data",
						Aliases: []string{"gh"},
						Usage:   "Include PR title, description",
					},
					&cli.IntFlag{
						Name:  "concurrency",
						Usage: "Sets how many inspectors will run concurrently",
					},
					&cli.StringSliceFlag{
						Name:    "inspector",
						Aliases: []string{"i"},
						Usage:   "Runs the provided inspector `script`",
					},
					// TODO add formatter override
				},
				Action: func(cctx *cli.Context) error {
					var in io.Reader

					fi, err := os.Stdin.Stat()
					if err != nil {
						panic(err)
					}
					if (fi.Mode() & os.ModeCharDevice) == 0 {
						in = os.Stdin
					} else if diff := cctx.String("diff"); diff != "" {
						f, err := os.Open(diff)
						if err != nil {
							return err
						}
						defer f.Close()
						in = f
					} else {
						if err := cli.ShowSubcommandHelp(cctx); err != nil {
							fmt.Println(err)
						}
						fmt.Printf("\n")
						return cli.Exit(color.New(color.FgRed).Sprint("No diff provided. Please provide a --diff or pass the diff via stdin."), 1)
					}

					// Setup root configuration
					customsConfig := &customs.Configuration{
						Concurrency: 1,
						Formatter:   prettyformat.New(os.Stdout),
						Inspectors:  map[string]string{},
					}
					err = applyConfig(cctx.String("config"), customsConfig)
					if err != nil {
						return err
					}

					populateInspectors(cctx, customsConfig)

					// config overrides
					if concurrency := cctx.Int("concurrency"); concurrency > 0 {
						customsConfig.Concurrency = concurrency
					}

					inspection, err := customs.NewInspection(customsConfig, in)
					if err != nil {
						color.New(color.FgRed).Println(err.Error())
						return cli.ShowSubcommandHelp(cctx)
					}
					err = populateGitHubData(cctx, customsConfig, inspection)
					if err != nil {
						return cli.Exit(err, 1)
					}

					if cctx.Bool("only-import-json") {
						out, err := inspection.ImportJSON()
						if err != nil {
							fmt.Printf("Could not return import JSON: %s\n", err)
						}

						fmt.Println(string(out))
						return nil
					}

					if len(customsConfig.Inspectors) == 0 {
						if err := cli.ShowSubcommandHelp(cctx); err != nil {
							fmt.Println(err)
						}
						fmt.Printf("\n")
						return cli.Exit(color.New(color.FgRed).Sprint("No inspectors were provided. Add one to customs.yaml or passed via --inspector"), 1)

					}

					err = inspection.Perform()
					if err != nil {
						return cli.Exit(color.New(color.FgRed).Sprintf("Customs inspection encountered an error: %s\n", err.Error()), 1)
					}

					color.New(color.FgGreen).Fprintf(os.Stderr, "Customs inspection passed!")

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

func applyConfig(configArg string, rootConfig *customs.Configuration) error {
	if configArg != "" {
		f, err := os.Open(configArg)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Could not open the provided config file: %s", err), 1)
		}
		defer f.Close()
		customs.ParseConfig(f, rootConfig, map[string]customs.Formatter{"pretty": prettyformat.New(os.Stdout)})

		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return cli.Exit("Could not get current working directory", 1)
	}
	rootDir, err := findGitDir(cwd)
	if err != nil && err != os.ErrNotExist {
		return cli.Exit(fmt.Sprintf("error when looking for root dir: %s", err), 1)
	}

	if err == os.ErrNotExist {
		return nil
	}

	configPath := filepath.Join(rootDir, "customs.yaml")
	if _, err := os.Stat(configPath); err == nil {
		f, err := os.Open(configPath)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Could not open the config file found in the root folder: %s", err), 1)
		}
		defer f.Close()

		customs.ParseConfig(f, rootConfig, map[string]customs.Formatter{"pretty": prettyformat.New(os.Stdout)})
	}

	return nil
}

func findGitDir(startDir string) (string, error) {
	dir := startDir
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", os.ErrNotExist
}

func populateGitHubData(cctx *cli.Context, c *customs.Configuration, i *customs.Inspection) error {
	if !cctx.Bool("include-pr-data") {
		return nil
	}

	// Ensure we have a token to fetch with
	token := os.Getenv("CUSTOMS_GITHUB_TOKEN")
	if token == "" {
		fmt.Fprint(os.Stderr, "CUSTOMS_GITHUB_TOKEN was not present so pull request information could not be fetched\n")
		return nil
	}

	// Get the owner and repo details so we can fetch from the API
	owner, repo, err := githelpers.NwoFromOrigin()
	if err != nil {
		return fmt.Errorf("Could not get owner and repo from git origin: %w", err)
	}

	// Get the most recent pushed SHA so we can fetch the PR details from GitHub
	sha, err := githelpers.UpstreamSha()
	if err != nil && err != githelpers.ErrNoPushedBranch {
		return fmt.Errorf("Could not get sha: %w", err)
	}

	gh := github.NewClient(token)

	return i.PopulatePullDetails(gh, owner, repo, sha)
}

func populateInspectors(cctx *cli.Context, c *customs.Configuration) {
	// include CLI defined inspectors
	if inspectors := cctx.StringSlice("inspector"); inspectors != nil {
		for _, inspector := range inspectors {
			c.Inspectors[inspector] = inspector
		}
	}
}
