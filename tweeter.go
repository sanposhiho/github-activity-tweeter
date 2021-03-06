package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
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

	user, resp, err := twiclient.Accounts.VerifyCredentials(&twitter.AccountVerifyParams{})
	if err != nil {
		return xerrors.Errorf("verify credentials: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return xerrors.New("response from twitter is not 200")
	}

	tweets, resp, err := twiclient.Timelines.UserTimeline(&twitter.UserTimelineParams{
		ScreenName: user.ScreenName,
	})
	if err != nil {
		return xerrors.Errorf("get recent tweet: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return xerrors.Errorf("statuscode is not 200, %d", resp.StatusCode)
	}

	for _, e := range events {
		if e.CreatedAt.Before(until) && e.CreatedAt.After(from) {
			msg, short, url, err := BuildMessage(e, generalconfig.GitHubUserName, generalconfig.ExcludeEvent, generalconfig.ExcludeRepoPattern)
			if err != nil {
				// ok to ignore, because this is not critical error.
				log.Println(err)
				continue
			}
			Tweet(twiclient, msg, short, url, tweets)
		}
	}
	return nil
}

// msgs is used for duplicate checking
var msgs = map[string]bool{}

// Tweet tweets given msg.
// If we cannot use msg because of the max number of text on one tweet, we will use short.
// It doesn't return error, but logging the error.
func Tweet(twiclient *twitter.Client, msg, short, url string, tweets []twitter.Tweet) {
	// message must be this format -- "some message || URL"
	mergedMsg := fmt.Sprintf("%s || %s", msg, url)
	log.Println("Try tweet: " + mergedMsg)
	// check if recently tweeted from message map
	if msgs[mergedMsg] {
		log.Println("recently tweeted, skipped")
		return
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
			break
		}
	}
	if hasRecentlyTweeted {
		log.Println("recently tweeted, skipped")
		return
	}

	// tweet
	_, resp, err := twiclient.Statuses.Update(mergedMsg, nil)
	if err != nil {
		var typed twitter.APIError
		ok := errors.As(err, &typed)
		if ok {
			isDuplicated := false
			for _, e := range typed.Errors {
				if e.Code == 186 {
					// too long msg

					log.Println("too long tweet, try it again with short message")
					// use short as msg and url as short.
					Tweet(twiclient, short, url, url, tweets)
					return
				}
				if e.Code == 187 {
					isDuplicated = true
				}
			}
			if isDuplicated {
				log.Println("duplicated tweet, skipped")
				return
			}
		}
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("statuscode is not 200, %d\n", resp.StatusCode)
	}

	log.Println("successfully tweeted")
	return
}

func NewTwitterClient(generalconfig *config.Config) *twitter.Client {
	config := oauth1.NewConfig(generalconfig.ConsumerKey, generalconfig.ConsumerSecret)
	token := oauth1.NewToken(generalconfig.AccessToken, generalconfig.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	return twitter.NewClient(httpClient)
}

// BuildMessage build a tweet message.
// It returns full message, short message, url and error.
func BuildMessage(e *github.Event, githubusername string, excludeEvent []string, excludeRepoPattern *regexp.Regexp) (string, string, string, error) {
	if excludeRepoPattern != nil && excludeRepoPattern.MatchString(*e.Repo.Name) {
		return "", "", "", xerrors.New("an event for the repository is excluded by config")
	}

	excludeEventMap := map[string]bool{}
	for _, ee := range excludeEvent {
		excludeEventMap[ee] = true
	}

	// TODO: Make it possible to set which events are tweeted.
	var msg string
	var shortmsg string
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
		if excludeEventMap["IssuesEvent"] {
			return "", "", "", xerrors.New("this event are excluded by config")
		}
		isu, err := e.ParsePayload()
		if err != nil {
			return "", "", "", xerrors.Errorf("parse payload: %w", err)
		}
		isuevent, ok := isu.(*github.IssuesEvent)
		if !ok {
			return "", "", "", xerrors.New("failed to convert to IssuesEvent")
		}
		if *isuevent.Action != "opened" {
			return "", "", "", xerrors.New("unsupported action on IssuesEvent")
		}

		msg = fmt.Sprintf("%s opened a issue in %s: %s", githubusername, *e.Repo.Name, *isuevent.Issue.Title)
		shortmsg = fmt.Sprintf("%s opened a issue in %s", githubusername, *e.Repo.Name)
		url = *isuevent.Issue.HTMLURL
	case "PullRequestEvent":
		if excludeEventMap["PullRequestEvent"] {
			return "", "", "", xerrors.New("this event are excluded by config")
		}
		pr, err := e.ParsePayload()
		if err != nil {
			return "", "", "", xerrors.Errorf("parse payload: %w", err)
		}
		prevent, ok := pr.(*github.PullRequestEvent)
		if !ok {
			return "", "", "", xerrors.New("failed to convert to PullRequestEvent")
		}
		if *prevent.Action != "opened" {
			return "", "", "", xerrors.New("unsupported action on PullRequestEvent")
		}

		msg = fmt.Sprintf("%s created a pull request in %s: %s", githubusername, *e.Repo.Name, *prevent.PullRequest.Title)
		shortmsg = fmt.Sprintf("%s created a pull request in %s", githubusername, *e.Repo.Name)
		url = *prevent.PullRequest.HTMLURL
	case "ReleaseEvent":
		if excludeEventMap["ReleaseEvent"] {
			return "", "", "", xerrors.New("this event are excluded by config")
		}
		pr, err := e.ParsePayload()
		if err != nil {
			return "", "", "", xerrors.Errorf("parse payload: %w", err)
		}
		typed, ok := pr.(*github.ReleaseEvent)
		if !ok {
			return "", "", "", xerrors.New("failed to convert to ReleaseEvent")
		}
		if *typed.Action != "created" && *typed.Action != "published" {
			return "", "", "", xerrors.New("unsupported action on PullRequestEvent")
		}

		msg = fmt.Sprintf("%s %s release %s of %s", githubusername, *typed.Action, *typed.Release.TagName, *e.Repo.Name)
		shortmsg = fmt.Sprintf("%s %s release %s of %s", githubusername, *typed.Action, *typed.Release.TagName, *e.Repo.Name)
		url = *typed.Release.HTMLURL
	case "RepositoryEvent":
		if excludeEventMap["RepositoryEvent"] {
			return "", "", "", xerrors.New("this event are excluded by config")
		}
		pay, err := e.ParsePayload()
		if err != nil {
			return "", "", "", xerrors.Errorf("parse payload: %w", err)
		}
		typed, ok := pay.(*github.RepositoryEvent)
		if !ok {
			return "", "", "", xerrors.New("failed to convert to RepositoryEvent")
		}

		if *typed.Action != "created" && *typed.Action != "publicized" {
			return "", "", "", xerrors.New("unsupported action on RepositoryEvent")
		}

		msg = fmt.Sprintf("%s %s repository %s: %s", githubusername, *typed.Action, *e.Repo.Name, *e.Repo.Description)
		shortmsg = fmt.Sprintf("%s %s repository %s", githubusername, *typed.Action, *e.Repo.Name)
		url = *typed.Repo.HTMLURL

	default:
		return "", "", "", xerrors.New("unsupported event")
	}

	return msg, shortmsg, url, nil
}
