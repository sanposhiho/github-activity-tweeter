package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
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

	userid := "timessanposhiho"
	tweets, resp, err := twiclient.Timelines.UserTimeline(&twitter.UserTimelineParams{
		ScreenName: userid,
	})
	if err != nil {
		return xerrors.Errorf("get recent tweet: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return xerrors.Errorf("statuscode is not 200, %d", resp.StatusCode)
	}

	// msgs is used for duplicate checking
	msgs := map[string]bool{}
	for _, e := range events {
		if e.CreatedAt.Before(until) && e.CreatedAt.After(from) {
			msg, url, err := BuildMessage(e, generalconfig.GitHubUserName)
			if err != nil {
				// ok to ignore, because this is not critical error.
				continue
			}

			// message must be this format -- "some message || URL"
			mergedMsg := fmt.Sprintf("%s || %s", msg, url)
			log.Println("Try tweet: " + mergedMsg)
			// check if recently tweeted from message map
			if msgs[mergedMsg] {
				log.Println("recently tweeted, skipped")
				continue
			}

			// make it true not to tweet again
			msgs[mergedMsg] = true

			// check if recently tweeted from user's timeline
			hasRecentlyTweeted := false
			for _, t := range tweets {
				trimedtweet := strings.Split(t.Text, "||")
				trimedmsg := strings.Split(mergedMsg, "||")
				if len(trimedmsg) != len(trimedtweet) {
					continue
				}

				isSame := true
				// last trimed list entity must be URL, and Twitter changes URL with their format(t.co), so it will always be different.
				for i := 0; i < len(trimedtweet)-1; i++ {
					if trimedmsg[i] != trimedtweet[i] {
						isSame = false
					}
				}
				if isSame {
					hasRecentlyTweeted = true
				}
			}
			if hasRecentlyTweeted {
				log.Println("recently tweeted, skipped")
				continue
			}

			// tweet
			_, resp, err = twiclient.Statuses.Update(mergedMsg, nil)
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

func BuildMessage(e *github.Event, githubusername string) (string, string, error) {
	// TODO: Make it possible to set which events are tweeted.
	var msg string
	var url string
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
	//			case "PullRequestReviewEvent":
	//			case "PullRequestReviewCommentEvent":
	//			case "PullRequestTargetEvent":
	//			case "PushEvent":
	//			case "RepositoryDispatchEvent":
	//			case "RepositoryVulnerabilityAlertEvent":
	//			case "StatusEvent":
	//			case "TeamEvent":
	//			case "TeamAddEvent":
	//			case "UserEvent":
	//			case "WatchEvent":
	//			case "WorkflowDispatchEvent":
	//			case "WorkflowRunEvent":
	case "IssuesEvent":
		isu, err := e.ParsePayload()
		if err != nil {
			return "", "", xerrors.Errorf("parse payload: %w", err)
		}
		isuevent, ok := isu.(*github.IssuesEvent)
		if !ok {
			return "", "", xerrors.New("failed to convert to IssuesEvent")
		}
		if *isuevent.Action != "opened" {
			return "", "", xerrors.New("unsupported action on IssuesEvent")
		}

		msg = fmt.Sprintf("%s opened a issue in %s", githubusername, *e.Repo.Name)
		url = *isuevent.Issue.HTMLURL
	case "PullRequestEvent":
		pr, err := e.ParsePayload()
		if err != nil {
			return "", "", xerrors.Errorf("parse payload: %w", err)
		}
		prevent, ok := pr.(*github.PullRequestEvent)
		if !ok {
			return "", "", xerrors.New("failed to convert to PullRequestEvent")
		}
		if *prevent.Action != "opened" {
			return "", "", xerrors.New("unsupported action on PullRequestEvent")
		}

		msg = fmt.Sprintf("%s created a pull request in %s", githubusername, *e.Repo.Name)
		url = *prevent.PullRequest.HTMLURL
	case "ReleaseEvent":
		pr, err := e.ParsePayload()
		if err != nil {
			return "", "", xerrors.Errorf("parse payload: %w", err)
		}
		typed, ok := pr.(*github.ReleaseEvent)
		if !ok {
			return "", "", xerrors.New("failed to convert to ReleaseEvent")
		}
		if *typed.Action != "created" && *typed.Action != "published" {
			return "", "", xerrors.New("unsupported action on PullRequestEvent")
		}

		msg = fmt.Sprintf("%s %s release %s of %s", githubusername, *typed.Action, *typed.Release.TagName, *e.Repo.Name)
		url = *typed.Release.HTMLURL
	case "RepositoryEvent":
		pay, err := e.ParsePayload()
		if err != nil {
			return "", "", xerrors.Errorf("parse payload: %w", err)
		}
		typed, ok := pay.(*github.RepositoryEvent)
		if !ok {
			return "", "", xerrors.New("failed to convert to RepositoryEvent")
		}

		if *typed.Action != "created" && *typed.Action != "publicized" {
			return "", "", xerrors.New("unsupported action on RepositoryEvent")
		}

		msg = fmt.Sprintf("%s %s repository %s", githubusername, *typed.Action, *e.Repo.Name)
		url = *typed.Repo.HTMLURL

	default:
		return "", "", xerrors.New("unsupported event")
	}

	return msg, url, nil
}
