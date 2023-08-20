package scan

import (
	"context"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/shurcooL/githubv4"
	"github.com/urfave/cli"
	"golang.org/x/oauth2"

	"github.com/nais/gh-org-stats/pkg/config"
)

func ScanCommand() cli.Command {
	return cli.Command{
		Name:  "scan",
		Usage: "list all repositories for the given organization",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:     "org",
				Usage:    "the name of the organization",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			ctx := context.Background()

			cfg, err := config.NewConfig(c)
			if err != nil {
				return err
			}

			src := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: cfg.GitHubToken},
			)
			httpClient := oauth2.NewClient(ctx, src)

			client := githubv4.NewClient(httpClient)

			var q query
			variables := map[string]interface{}{
				"org":   githubv4.String(c.String("org")),
				"limit": githubv4.Int(10),
			}

			err = client.Query(ctx, &q, variables)
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Created At", "Fork Count", "Stargazer Count", "License", "PR Count", "Default Branch", "Status", "Last Commit"})
			table.SetAutoWrapText(false)

			for _, repo := range q.Organization.Repositories.Nodes {
				var status string
				var lastCommit string

				if repo.DefaultBranchRef.Target.Commit.Status.State == "SUCCESS" {
					status = "✔"
				} else {
					status = "✘"
				}

				if len(repo.DefaultBranchRef.Target.Commit.History.Edges) > 0 {
					lastCommit = repo.DefaultBranchRef.Target.Commit.History.Edges[0].Node.CommittedDate.Format("2006-01-02 15:04:05")
				}

				table.Append([]string{
					string(repo.Name),
					repo.CreatedAt.Format("2006-01-02"),
					fmt.Sprintf("%d", repo.ForkCount),
					fmt.Sprintf("%d", repo.StargazerCount),
					string(repo.LicenseInfo.Name),
					fmt.Sprintf("%d", repo.PullRequests.TotalCount),
					string(repo.DefaultBranchRef.Name),
					status,
					lastCommit,
				})
			}

			table.Render()

			fmt.Printf("\nRate limit: %d/%d\n", q.RateLimit.Remaining, q.RateLimit.Limit)

			return nil

		},
	}
}
