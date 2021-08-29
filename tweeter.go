package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dghubble/go-twitter/twitter"

	"golang.org/x/xerrors"

	"github.com/sanposhiho/github-activity-tweeter/config"

	"github.com/dghubble/oauth1"
	"github.com/google/go-github/v38/github"
)

func main() {
	if err := tweet(); err != nil {
		log.Fatalf("failed to tweet: %+v", err)
	}
}

func tweet() error {
	ctx := context.Background()
	generalconfig, err := config.NewConfig()
	if err != nil {
		return xerrors.Errorf("get configuration: %w", err)
	}

	until := time.Now()
	from := until.Add(-generalconfig.Interval)

	twiclient := NewTwitterClient(generalconfig)

	client := github.NewClient(nil)
	// TODO: considering paging
	events, res, err := client.Activity.ListEventsPerformedByUser(ctx, generalconfig.GitHubUserName, true, &github.ListOptions{PerPage: 100})
	if err != nil {
		return xerrors.Errorf("list events performed by user: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return xerrors.New("response from github is not 200")
	}
	for _, e := range events {
		if e.CreatedAt.Before(until) && e.CreatedAt.After(from) {
			// TODO: Make it possible to set which events are tweeted.
			var msg string
			switch *e.Type {
			//			case "CheckRunEvent":
			//			case "CheckSuiteEvent":
			//			case "CommitCommentEvent":
			//			case "ContentReferenceEvent":
			//			case "CreateEvent":
			//			case "DeleteEvent":
			//			case "DeployKeyEvent":
			//			case "DeploymentEvent":
			//			case "DeploymentStatusEvent":
			//			case "ForkEvent":
			//			case "GitHubAppAuthorizationEvent":
			//			case "GollumEvent":
			//			case "InstallationEvent":
			//			case "InstallationRepositoriesEvent":
			//			case "IssueCommentEvent":
			case "IssuesEvent":
				isu, err := e.ParsePayload()
				if err != nil {
					return xerrors.Errorf("parse payload: %w", err)
				}
				isuevent, ok := isu.(*github.IssuesEvent)
				if !ok {
					return xerrors.New("failed to convert to IssuesEvent")
				}

				msg = fmt.Sprintf("%s created a issue in %s || %s", generalconfig.GitHubUserName, *e.Repo.Name, *isuevent.Issue.HTMLURL)
			//			case "LabelEvent":
			//			case "MarketplacePurchaseEvent":
			//			case "MemberEvent":
			//			case "MembershipEvent":
			//			case "MetaEvent":
			//			case "MilestoneEvent":
			//			case "OrganizationEvent":
			//			case "OrgBlockEvent":
			//			case "PackageEvent":
			//			case "PageBuildEvent":
			//			case "PingEvent":
			//			case "ProjectEvent":
			//			case "ProjectCardEvent":
			//			case "ProjectColumnEvent":
			//			case "PublicEvent":
			case "PullRequestEvent":
				pr, err := e.ParsePayload()
				if err != nil {
					return xerrors.Errorf("parse payload: %w", err)
				}
				prevent, ok := pr.(*github.PullRequestEvent)
				if !ok {
					return xerrors.New("failed to convert to PullRequestEvent")
				}

				msg = fmt.Sprintf("%s created a pull request in %s || %s", generalconfig.GitHubUserName, *e.Repo.Name, *prevent.PullRequest.HTMLURL)
			//			case "PullRequestReviewEvent":
			//			case "PullRequestReviewCommentEvent":
			//			case "PullRequestTargetEvent":
			//			case "PushEvent":
			//			case "ReleaseEvent":
			case "RepositoryEvent":
				pay, err := e.ParsePayload()
				if err != nil {
					return xerrors.Errorf("parse payload: %w", err)
				}
				typed, ok := pay.(*github.RepositoryEvent)
				if !ok {
					return xerrors.New("failed to convert to RepositoryEvent")
				}

				if *typed.Action != "created" && *typed.Action != "publicized" {
					continue
				}

				msg = fmt.Sprintf("%s %s repository %s || %s", generalconfig.GitHubUserName, *typed.Action, *e.Repo.Name, *typed.Repo.HTMLURL)
				//			case "RepositoryDispatchEvent":
				//			case "RepositoryVulnerabilityAlertEvent":
				//			case "StarEvent":
				//			case "StatusEvent":
				//			case "TeamEvent":
				//			case "TeamAddEvent":
				//			case "UserEvent":
				//			case "WatchEvent":
				//			case "WorkflowDispatchEvent":
				//			case "WorkflowRunEvent":
			default:
				continue
			}
			log.Println("Try tweet: " + msg)
			_, resp, err := twiclient.Statuses.Update(msg, nil)
			if err != nil {
				var typed twitter.APIError
				ok := errors.As(err, &typed)
				if ok {
					isDuplicated := false
					for _, e := range typed.Errors {
						if e.Code == 187 {
							log.Println("duplicated tweet, skipped")
							isDuplicated = true
						}
					}
					if isDuplicated {
						continue
					}
				}

				return xerrors.Errorf("tweet: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Printf("statuscode is not 200, %d\n", resp.StatusCode)
			}
		}
	}

	return nil
}

func NewTwitterClient(generalconfig *config.Config) *twitter.Client {
	config := oauth1.NewConfig(generalconfig.ConsumerKey, generalconfig.ConsumerSecret)
	token := oauth1.NewToken(generalconfig.AccessToken, generalconfig.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	return twitter.NewClient(httpClient)
}
