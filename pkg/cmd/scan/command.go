package scan

import (
	"context"
	"fmt"
	"math"
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

func daysAgo(t time.Time) int {
	return int(time.Since(t).Hours() / 24)
}

func commitStatus(checks []checkSuit) (status string) {
	for _, check := range checks {
		if check.Node.Conclusion == "FAILURE" {
			status = "✘"
			break
		}

		if check.Node.Conclusion == "SUCCESS" {
			status = "✔"
		}
	}

	return
}

func lastCommit(c commit) string {
	return fmt.Sprintf("%d days ago", daysAgo(c.CommittedDate.Time))
}

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
			cli.IntFlag{
				Name:  "limit",
				Usage: "limit the number of repositories to scan",
				Value: 0,
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
				"limit":  githubv4.Int(math.Min(float64(queryLimit), float64(c.Int("limit")))),
				"cursor": (*githubv4.String)(nil), // Null after argument to get first page.
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Created At", "Last commit", "Stars / Forks", "License", "Language", "Default Branch", "Status"})
			table.SetAutoWrapText(false)
			table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})

			var count int

			for {
				var q query

				retries := 5
				for retries > 0 {
					cctx, cancel := context.WithTimeout(ctx, queryTimeout)
					defer cancel()

					err = client.Query(cctx, &q, variables)
					if err != nil {
						fmt.Println(err.Error())

						sleep := time.Duration(queryRetryFactor*retries) * time.Second

						fmt.Printf("Error: %v, Retrying in %d seconds...\n", err, sleep)
						time.Sleep(sleep)
						retries--
						continue
					}

					break
				}

				for _, repo := range q.Organization.Repositories.Nodes {
					count++

					status := commitStatus(repo.DefaultBranchRef.Target.Commit.CheckSuit.Edges)
					lastCommit := lastCommit(repo.DefaultBranchRef.Target.Commit)

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

				// If there's no next page, we're done.
				if !q.Organization.Repositories.PageInfo.HasNextPage {
					fmt.Printf("Rate limit: %d/%d\n", q.RateLimit.Remaining, q.RateLimit.Limit)
					break
				}

				// If we've reached the limit, we're done.
				if c.Int("limit") > 0 && count >= c.Int("limit") {
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
