package main

import (
	"os"

	"github.com/nais/gh-org-stats/pkg/cmd/contributions"
	"github.com/nais/gh-org-stats/pkg/cmd/scan"
	"github.com/nais/gh-org-stats/pkg/config"
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

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:     "github-token",
			Usage:    "GitHub token",
			EnvVar:   "GITHUB_TOKEN",
			Required: true,
		},
		cli.StringFlag{
			Name:     "google-project-id",
			Usage:    "Google Cloud project ID",
			EnvVar:   "GOOGLE_PROJECT_ID",
			Required: true,
		},
		cli.StringFlag{
			Name:     "bigquery-dataset-id",
			Usage:    "BigQuery dataset ID",
			EnvVar:   "BIGQUERY_DATASET_ID",
			Required: true,
		},
		cli.IntFlag{
			Name:   "default-query-timeout",
			Usage:  "default query timeout in seconds",
			Value:  20,
			Hidden: true,
			EnvVar: "DEFAULT_QUERY_TIMEOUT",
		},
	}

	app.Commands = []cli.Command{
		scan.ScanCommand(),
		contributions.ContributionsCommand(),
	}

	err = app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
