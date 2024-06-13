package contributions

import "github.com/shurcooL/githubv4"

type commit struct {
	Oid           githubv4.GitObjectID
	CommittedDate githubv4.DateTime
	Message       githubv4.String
	Status        struct{ State githubv4.String }
	CheckSuit     struct {
		TotalCount githubv4.Int
		Edges      []checkSuit
	} `graphql:"checkSuites(first: 10)"`
}

type checkSuit struct {
	Node struct {
		Status     githubv4.String
		Conclusion githubv4.String
	}
}

type contributionsByRepo struct {
	Repository repo
}

type repo struct {
	Name          githubv4.String
	NameWithOwner githubv4.String
	Owner         struct {
		Login githubv4.String
	}
}

type query struct {
	Organization struct {
		Members struct {
			Nodes []struct {
				Login                   githubv4.String
				ContributionsCollection struct {
					IssueContributionsByRepository       []contributionsByRepo `graphql:"issueContributionsByRepository(maxRepositories:100)"`
					PullRequestContributionsByRepository []contributionsByRepo `graphql:"pullRequestContributionsByRepository(maxRepositories:100)"`
				} `graphql:"contributionsCollection(from: $contribFrom, to: $contribTo)"`
			}
			TotalCount githubv4.Int
			PageInfo   struct {
				EndCursor   githubv4.String
				HasNextPage githubv4.Boolean
			}
		} `graphql:"membersWithRole(first: $limit, after: $cursor)"`
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
