package scan

import "github.com/shurcooL/githubv4"

type query struct {
	Organization struct {
		Repositories struct {
			Nodes []struct {
				Name             githubv4.String
				Description      githubv4.String
				CreatedAt        githubv4.DateTime
				ForkCount        githubv4.Int
				StargazerCount   githubv4.Int
				LicenseInfo      struct{ Name githubv4.String }
				PullRequests     struct{ TotalCount githubv4.Int }
				DefaultBranchRef struct {
					Name   githubv4.String
					Target struct {
						Commit struct {
							Status  struct{ State githubv4.String }
							History struct {
								TotalCount githubv4.Int
								Edges      []struct {
									Node struct {
										Oid           githubv4.GitObjectID
										Message       githubv4.String
										CommittedDate githubv4.DateTime
									}
								}
							} `graphql:"history(first: 1)"`
						} `graphql:"... on Commit"`
					}
				}

				PrimaryLanguage struct{ Name githubv4.String }
				Languages       struct {
					Edges []struct {
						Node struct{ Name githubv4.String }
					}
				} `graphql:"languages(first: 5)"`

				IsPrivate  githubv4.Boolean
				IsArchived githubv4.Boolean
				IsFork     githubv4.Boolean
			}
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage githubv4.Boolean
			}
		} `graphql:"repositories(orderBy: {field: CREATED_AT, direction: DESC}, privacy: PUBLIC, isArchived:false, first: $limit)"`
	} `graphql:"organization(login: $org)"`
	Viewer struct {
		Login      githubv4.String
		CreatedAt  githubv4.DateTime
		ID         githubv4.ID
		DatabaseID githubv4.Int
	}
	RateLimit struct {
		Cost      githubv4.Int
		Limit     githubv4.Int
		Remaining githubv4.Int
		ResetAt   githubv4.DateTime
	}
}
