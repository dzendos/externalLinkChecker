package repo_parser

import (
	"context"
	"externalLinkChecker/internal/config"
	"fmt"
	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
	"log"
	"time"
)

type commentParser interface {
	HandleCommentsByURL(repository, url, typename string, createdAt, closedAt *time.Time)
}

type db interface {
	InsertNewRepo(repository, owner, url, lang string)
	IncrementIssue(delta int)
	IncrementPull(delta int)
}

type RepoParser struct {
	gh *github.Client
	cp commentParser
	db db
}

func New(ctx context.Context, cfg *config.Config, cp commentParser, db db) *RepoParser {
	// GH Client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Github.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	gh := github.NewClient(tc)

	return &RepoParser{
		gh: gh,
		cp: cp,
		db: db,
	}
}

func (r *RepoParser) getRepos(ctx context.Context) []*github.Repository {
	var repos []*github.Repository

	fetchRepositories := func(orgName string) {
		limit := 1000
		opt := &github.RepositoryListByOrgOptions{
			ListOptions: github.ListOptions{PerPage: 100},
		}

		var totalCount int
		for {
			rps, resp, err := r.gh.Repositories.ListByOrg(ctx, orgName, opt)
			if err != nil {
				log.Fatalf("Error fetching repositories for %s: %v", orgName, err)
			}

			repos = append(repos, rps...)

			totalCount += len(rps)
			if totalCount >= limit {
				break
			}

			if resp.NextPage == 0 {
				break
			}

			opt.Page = resp.NextPage
		}
	}

	fetchRepositories("microsoft")
	fetchRepositories("google")
	fetchRepositories("yandex")
	fetchRepositories("aws")
	fetchRepositories("docker")
	fetchRepositories("apple")
	fetchRepositories("openai")

	return repos
}

func (r *RepoParser) getRepoPullRequests(ctx context.Context, repo *github.Repository) []*github.PullRequest {
	owner := repo.GetOwner().GetLogin()
	name := repo.GetName()

	pullRequests, _, err := r.gh.PullRequests.List(ctx, owner, name, nil)
	if err != nil {
		log.Fatalf("Error fetching pull requests: %v\n", err)
	}

	fmt.Printf("Pull Requests for %s/%s; count %v:\n", owner, name, len(pullRequests))
	return pullRequests
}

func (r *RepoParser) getIssues(ctx context.Context, repo *github.Repository) []*github.Issue {
	owner := repo.GetOwner().GetLogin()
	name := repo.GetName()

	issues, _, err := r.gh.Issues.ListByRepo(ctx, owner, name, nil)
	if err != nil {
		log.Fatalf("Error fetching pull requests: %v\n", err)
	}

	fmt.Printf("Issues for %s/%s; count %v:\n", owner, name, len(issues))
	return issues
}

func (r *RepoParser) Run(ctx context.Context) {
	repos := r.getRepos(ctx)
	log.Println("total number of repos:", len(repos))

	for _, repo := range repos {
		r.db.InsertNewRepo(repo.GetName(), repo.GetOwner().GetLogin(), repo.GetHTMLURL(), repo.GetLanguage())
		//pullRequests := r.getRepoPullRequests(ctx, repo)
		//for _, pr := range pullRequests {
		//	r.cp.HandleCommentsByURL(repo.GetName(), pr.GetHTMLURL(), "pull", pr.CreatedAt.GetTime(), pr.ClosedAt.GetTime())
		//	r.db.IncrementPull(1)
		//}
		//
		//issues := r.getIssues(ctx, repo)
		//for _, issue := range issues {
		//	r.cp.HandleCommentsByURL(repo.GetName(), issue.GetHTMLURL(), "issue", issue.CreatedAt.GetTime(), issue.ClosedAt.GetTime())
		//	r.db.IncrementIssue(1)
		//}
	}
}
