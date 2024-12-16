package main

import (
	"os"

	"github.com/blakewilliams/customs/cli"
)

func main() {
	app := cli.New()
	app.Run(os.Args)
}
