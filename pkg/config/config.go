package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

type Config struct {
	GitHubToken string
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
