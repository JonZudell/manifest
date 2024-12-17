package main

import (
	"os"

	"github.com/blakewilliams/manifest/cli"
)

func main() {
	app := cli.New()
	app.Run(os.Args)
}
