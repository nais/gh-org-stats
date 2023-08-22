package scan

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/shurcooL/githubv4"
	"github.com/urfave/cli"
	"golang.org/x/oauth2"

	"github.com/nais/gh-org-stats/pkg/config"
)

var (
	queryLimit       = 100
	queryTimeout     = 20 * time.Second
	queryRetryFactor = 2
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

			cctx, cancel := context.WithTimeout(ctx, queryTimeout)
			defer cancel()
			httpClient := oauth2.NewClient(cctx, src)

			client := githubv4.NewClient(httpClient)

			variables := map[string]interface{}{
				"org":    githubv4.String(c.String("org")),
				"limit":  githubv4.Int(queryLimit),
				"cursor": (*githubv4.String)(nil), // Null after argument to get first page.
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Created At", "Last commit", "Stars / Forks", "License", "Language", "Default Branch", "Status"})
			table.SetAutoWrapText(false)

			for {
				var q query

				retries := 5
				for retries > 0 {
					cctx, cancel := context.WithTimeout(ctx, queryTimeout)
					defer cancel()

					err = client.Query(cctx, &q, variables)
					if err != nil {
						fmt.Println(err.Error())

						sleep := time.Duration(queryRetryFactor) * 5 * time.Second

						fmt.Printf("Error: %v, Retrying in %d seconds...\n", err, sleep)
						time.Sleep(sleep)
						retries--
						continue
					}

					break
				}

				for _, repo := range q.Organization.Repositories.Nodes {
					var status string
					var lastCommit string

					status = ""
					for _, edge := range repo.DefaultBranchRef.Target.Commit.CheckSuit.Edges {
						if edge.Node.Conclusion == "FAILURE" {
							status = "✘"
							break
						}

						if edge.Node.Conclusion == "SUCCESS" {
							status = "✔"
						}
					}

					if len(repo.DefaultBranchRef.Target.Commit.History.Edges) > 0 {
						lastCommit = repo.DefaultBranchRef.Target.Commit.History.Edges[0].Node.CommittedDate.Format("2006-01-02")
					}

					table.Append([]string{
						string(repo.Name),
						repo.CreatedAt.Format("2006-01-02"),
						lastCommit,
						fmt.Sprintf("%d/%d", repo.StargazerCount, repo.ForkCount),
						string(repo.LicenseInfo.Name),
						string(repo.PrimaryLanguage.Name),
						string(repo.DefaultBranchRef.Name),
						status,
					})
				}

				if !q.Organization.Repositories.PageInfo.HasNextPage {
					fmt.Printf("Rate limit: %d/%d\n", q.RateLimit.Remaining, q.RateLimit.Limit)
					break
				}

				fmt.Printf("Cursor: %s, RateLimit: %d\n", q.Organization.Repositories.PageInfo.EndCursor, q.RateLimit.Remaining)

				variables["cursor"] = githubv4.NewString(q.Organization.Repositories.PageInfo.EndCursor)
			}

			table.Render()

			return nil
		},
	}
}
