# github-activity-tweeter

tweet your activity on GitHub. It is an useful tool to use on GitHub Action.

# Getting started

## Install

### Go version < 1.16

```
GO111MODULE=on go get github.com/sanposhiho/github-activity-tweeter
```

### Go 1.16+

```
go install github.com/sanposhiho/github-activity-tweeter@latest
```

## Configure

You need to get tokens for twitter. see: https://developer.twitter.com/en/apply-for-access

And you need to pass these environment variables to configure.

```
USER_NAME(string): github user name you want to get activity
INTERVAL(duration): How far back in time to get activities. (like 1s, 2m, 3h...)
ACCESS_TOKEN_SECRET(string): access token secret for twitter.
ACCESS_TOKEN(string): access token for twitter.
CONSUMER_SECRET(string): consumer secret for twitter.
CONSUMER_KEY(string): consumer key for twitter.
```

And, you can customize with these environment variables.

```
EXCLUDE_EVENT(string): you can exclude some event type. see below about event types. And you can pass multiple event with , separated values.
EXCLUDE_REPO(regexp): you can exclude events of some repository. 
```

## Run

```
$ github-activity-tweeter
```

**TIPs** 
- It checks your timeline and don't tweet duplicated one.
- It tweets short message if generated message is too long.

## Run on GitHub Action

Please see [tweet.yaml](.github/workflows/tweet.yaml). 

You only have to folk this repo and add needed environment variables to repository secrets.

## What does this tweet about?

- User opened an issue.(eventtype: `IssuesEvent`)
- User opened a pull request.(eventtype: `PullRequestEvent`)
- User created/published a release.(eventtype: `ReleaseEvent`)
- User create/publicized a repository.(eventtype: `RepositoryEvent`)
