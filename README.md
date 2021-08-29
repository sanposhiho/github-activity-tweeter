# github-activity-tweeter

tweet your activity on GitHub. 

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

You need to pass some env to configure.

```
USER_NAME: github user name you want to get activity
INTERVAL: How far back in time to get activities.
ACCESS_TOKEN_SECRET: access token secret for twitter.
ACCESS_TOKEN: access token for twitter.
CONSUMER_SECRET: consumer secret for twitter.
CONSUMER_KEY: consumer key for twitter.
```

## Run

```
$ github-activity-tweeter
```

And I run it on GitHub Action. Please see [tweet.yaml](.github/workflows/tweet.yaml) for details.

### Note

- duplicated tweet will be skipped.