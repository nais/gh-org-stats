package contributions

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/olekukonko/tablewriter"
	"github.com/shurcooL/githubv4"
	"github.com/urfave/cli"

	"github.com/nais/gh-org-stats/pkg/config"
)

var (
	queryLimit       = 10
	queryTimeout     = 20 * time.Second
	queryRetryFactor = 2
)

func outsideContributions(r []contributionsByRepo, user githubv4.String) (repos []string, orgs []string) {
	for _, repo := range r {
		if repo.Repository.Owner.Login != "navikt" && repo.Repository.Owner.Login != "nais" && repo.Repository.Owner.Login != user {
			repos = addUnique(repos, string(repo.Repository.NameWithOwner))
			orgs = addUnique(orgs, string(repo.Repository.Owner.Login))
		}
	}

	return repos, orgs
}

func addUniques(slice []string, items []string) []string {
	for _, item := range items {
		slice = addUnique(slice, item)
	}

	return slice
}

func addUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

func ContributionsCommand() cli.Command {
	return cli.Command{
		Name:  "contributions",
		Usage: "list all open source contributions for members of an organization",
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

			ghclient, gccancel := config.NewGitHubClient(c, ctx)
			defer gccancel()

			bqclient, bqcancel, err := config.NewBigQueryClient(c, ctx)
			if err != nil {
				return err
			}
			defer bqcancel()

			bqschema, err := bigquery.InferSchema(query{})
			if err != nil {
				return err
			}

			bqtable := bqclient.Dataset(c.GlobalString("bigquery-dataset-id")).Table("contributions")
			if err := bqtable.Create(ctx, &bigquery.TableMetadata{Schema: bqschema}); err != nil {
				// check if invalid schema
				if err.Error() != "googleapi: Error 409: Invalid schema update. Field login is missing., invalid" {
					return err
				}

				if _, err := bqtable.Update(ctx, bigquery.TableMetadataToUpdate{Schema: bqschema}, ""); err != nil {
					return err
				}
			}

			variables := map[string]interface{}{
				"org":    githubv4.String(c.String("org")),
				"limit":  githubv4.Int(math.Min(float64(queryLimit), float64(c.Int("limit")))),
				"cursor": (*githubv4.String)(nil), // Null after argument to get first page.
				"contribFrom": githubv4.DateTime{
					Time: time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.Now().Location()),
				},
				"contribTo": githubv4.DateTime{
					Time: time.Now(),
				},
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Login", "Issues", "Pull Requests"})
			table.SetAutoWrapText(false)

			var count int
			var repos []string
			var orgs []string

			for {
				var q query

				retries := 5
				for retries > 0 {
					cctx, cancel := context.WithTimeout(ctx, queryTimeout)
					defer cancel()

					err := ghclient.Query(cctx, &q, variables)
					if err != nil {
						fmt.Println(err.Error())

						sleep := time.Duration(queryRetryFactor*retries) * time.Second

						fmt.Printf("Error: %v, Retrying in %f seconds...\n", err, sleep.Seconds())
						time.Sleep(sleep)
						retries--
						continue
					}

					break
				}

				for _, member := range q.Organization.Members.Nodes {
					count++

					issueRepos, issueOrgs := outsideContributions(member.ContributionsCollection.IssueContributionsByRepository, member.Login)
					prRepos, prOrgs := outsideContributions(member.ContributionsCollection.PullRequestContributionsByRepository, member.Login)

					if len(issueRepos) == 0 && len(prRepos) == 0 {
						continue
					}

					repos = addUniques(repos, issueRepos)
					repos = addUniques(repos, prRepos)

					orgs = addUniques(orgs, issueOrgs)
					orgs = addUniques(orgs, prOrgs)

					table.Append([]string{
						string(member.Login),
						fmt.Sprintf("%d", len(issueRepos)),
						fmt.Sprintf("%d", len(prRepos)),
					})
				}

				// If there's no next page, we're done.
				if !q.Organization.Members.PageInfo.HasNextPage {
					fmt.Printf("Rate limit: %d/%d\n", q.RateLimit.Remaining, q.RateLimit.Limit)
					break
				}

				// If we've reached the limit, we're done.
				if c.Int("limit") > 0 && count >= c.Int("limit") {
					fmt.Printf("Rate limit: %d/%d\n", q.RateLimit.Remaining, q.RateLimit.Limit)
					break
				}

				fmt.Printf("Cursor: %s, RateLimit: %d\n", q.Organization.Members.PageInfo.EndCursor, q.RateLimit.Remaining)
				variables["cursor"] = githubv4.NewString(q.Organization.Members.PageInfo.EndCursor)
			}

			table.Render()

			fmt.Printf("Total number of repositories: %d\n", len(repos))
			fmt.Printf("Total number of organizations: %d\n", len(orgs))
			fmt.Printf("Organizations:\n")
			for _, org := range orgs {
				fmt.Printf("%s\n", org)
			}

			return nil
		},
	}
}
