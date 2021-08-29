# github-activity-tweeter

tweet your activity on GitHub. 

It is useful to use it on GitHub Action.

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

You need to pass these environment variables to configure.

```
USER_NAME: github user name you want to get activity
INTERVAL: How far back in time to get activities. (like 1s, 2m, 3h...)
ACCESS_TOKEN_SECRET: access token secret for twitter.
ACCESS_TOKEN: access token for twitter.
CONSUMER_SECRET: consumer secret for twitter.
CONSUMER_KEY: consumer key for twitter.
```

And, you can customize with these environment variables.

```
EXCLUDE_EVENT: you can exclude some event type. see below about event types. And you can pass multiple event with , separated values.
EXCLUDE_REPO: you can exclude events of some repository. You can use regexp. 
```

## Run

```
$ github-activity-tweeter
```

And I run it on GitHub Action. Please see [tweet.yaml](.github/workflows/tweet.yaml) for details.

## What does this tweet about?

- User opened an issue.(eventtype: `IssuesEvent`)
- User opened a pull request.(eventtype: `PullRequestEvent`)
- User created/published a release.(eventtype: `ReleaseEvent`)
- User create/publicized a repository.(eventtype: `RepositoryEvent`)

## Useful feature

- check your timeline and don't tweet duplicated one
