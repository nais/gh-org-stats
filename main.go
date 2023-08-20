package main

import (
	"os"

	"github.com/nais/gh-org-stats/pkg/config"
	"github.com/nais/gh-org-stats/pkg/scan"
	"github.com/urfave/cli"
)

func main() {
	err := config.LoadEnv()
	if err != nil {
		panic(err)
	}

	app := cli.NewApp()
	app.Name = "gh-org-stats"
	app.Usage = "a tool for gathering statistics about an organization's GitHub repositories"
	app.Version = "1.0.0"

	app.Commands = []cli.Command{
		scan.ScanCommand(),
	}

	err = app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
