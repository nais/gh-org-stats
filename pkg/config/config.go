package config

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/shurcooL/githubv4"
	"github.com/urfave/cli"
	"golang.org/x/oauth2"

	"cloud.google.com/go/bigquery"
)

type Config struct {
	GitHubToken       string
	GoogleProjectID   string
	BigqueryDatasetID string
}

func NewGitHubClient(c *cli.Context, ctx context.Context) (*githubv4.Client, context.CancelFunc) {
	cctx, cancel := context.WithTimeout(ctx, time.Duration(c.GlobalInt("default-query-timeout"))*time.Second)

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.GlobalString("github-token")},
	)

	httpClient := oauth2.NewClient(cctx, src)

	return githubv4.NewClient(httpClient), cancel
}

func NewBigQueryClient(c *cli.Context, ctx context.Context) (*bigquery.Client, context.CancelFunc, error) {
	cctx, cancel := context.WithTimeout(ctx, time.Duration(c.GlobalInt("default-query-timeout"))*time.Second)

	client, err := bigquery.NewClient(cctx, c.GlobalString("google-project-id"))
	if err != nil {
		return nil, cancel, err
	}

	return client, cancel, nil
}

func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	return nil
}

func NewConfig(c *cli.Context) (*Config, error) {
	cfg := &Config{}

	// Read GitHub token from environment variable or command-line flag
	if c.GlobalIsSet("github-token") {
		cfg.GitHubToken = c.GlobalString("github-token")
	} else if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		cfg.GitHubToken = token
	} else {
		return nil, errors.New("GitHub token not set")
	}

	return cfg, nil
}
