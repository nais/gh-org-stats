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
)

type Config struct {
	GitHubToken string
}

func (c *Config) GitHubClient(ctx context.Context, queryTimeout time.Duration) (context.CancelFunc, *githubv4.Client) {
	cctx, cancel := context.WithTimeout(ctx, queryTimeout)

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.GitHubToken},
	)

	httpClient := oauth2.NewClient(cctx, src)

	return cancel, githubv4.NewClient(httpClient)
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
