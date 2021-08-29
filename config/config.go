package config

import (
	"errors"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/xerrors"
)

// ErrEmptyEnv represents the required environment variable don't exist.
var ErrEmptyEnv = errors.New("env is needed, but empty")

// Config is configuration for simulator.
type Config struct {
	Interval           time.Duration
	GitHubUserName     string
	ExcludeRepoPattern *regexp.Regexp
	ExcludeEvent       []string

	// For Twitter
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

// NewConfig gets some settings from environment variables.
func NewConfig() (*Config, error) {
	githubusername, err := getGitHubUserName()
	if err != nil {
		return nil, xerrors.Errorf("get GitHubUserName: %w", err)
	}
	interval, err := getInterval()
	if err != nil {
		return nil, xerrors.Errorf("get interval: %w", err)
	}
	consumerkey, err := getConsumerKey()
	if err != nil {
		return nil, xerrors.Errorf("get ConsumerKey")
	}
	consec, err := getConsumerSecret()
	if err != nil {
		return nil, xerrors.Errorf("get ConsumerSecret")
	}
	at, err := getAccessToken()
	if err != nil {
		return nil, xerrors.Errorf("get AccessToken")
	}
	ats, err := getAccessTokenSecret()
	if err != nil {
		return nil, xerrors.Errorf("get AccessTokenSecret")
	}
	excludeRepo, err := getExcludeRepoPattern()
	if err != nil {
		return nil, xerrors.Errorf("get ExcludeRepoPattern")
	}
	excludeEvent, err := getExcludeEvent()
	if err != nil {
		return nil, xerrors.Errorf("get ExcludeEvent")
	}

	return &Config{
		GitHubUserName:     githubusername,
		Interval:           interval,
		ConsumerKey:        consumerkey,
		ConsumerSecret:     consec,
		AccessToken:        at,
		AccessTokenSecret:  ats,
		ExcludeEvent:       excludeEvent,
		ExcludeRepoPattern: excludeRepo,
	}, nil
}

func getExcludeRepoPattern() (*regexp.Regexp, error) {
	e := os.Getenv("EXCLUDE_REPO")
	if e == "" {
		return nil, nil
	}
	r, err := regexp.Compile(e)
	if err != nil {
		return nil, xerrors.Errorf("compile regexp on EXCLUDE_REPO: %w", err)
	}

	return r, nil
}

func getExcludeEvent() ([]string, error) {
	e := os.Getenv("EXCLUDE_EVENT")
	if e == "" {
		return nil, nil
	}
	splited := strings.Split(e, ",")

	return splited, nil
}

func getGitHubUserName() (string, error) {
	e := os.Getenv("USER_NAME")
	if e == "" {
		return "", xerrors.Errorf("get USER_NAME from env: %w", ErrEmptyEnv)
	}

	return e, nil
}

func getInterval() (time.Duration, error) {
	e := os.Getenv("INTERVAL")
	if e == "" {
		return 0, xerrors.Errorf("get INTERVAL from env: %w", ErrEmptyEnv)
	}

	t, err := time.ParseDuration(e)
	if err != nil {
		return 0, xerrors.Errorf("parse duration: %w", err)
	}

	return t, nil
}

func getAccessTokenSecret() (string, error) {
	e := os.Getenv("ACCESS_TOKEN_SECRET")
	if e == "" {
		return "", xerrors.Errorf("get ACCESS_TOKEN_SECRET from env: %w", ErrEmptyEnv)
	}

	return e, nil
}
func getAccessToken() (string, error) {
	e := os.Getenv("ACCESS_TOKEN")
	if e == "" {
		return "", xerrors.Errorf("get ACCESS_TOKEN from env: %w", ErrEmptyEnv)
	}

	return e, nil
}
func getConsumerSecret() (string, error) {
	e := os.Getenv("CONSUMER_SECRET")
	if e == "" {
		return "", xerrors.Errorf("get CONSUMER_SECRET from env: %w", ErrEmptyEnv)
	}

	return e, nil
}
func getConsumerKey() (string, error) {
	e := os.Getenv("CONSUMER_KEY")
	if e == "" {
		return "", xerrors.Errorf("get CONSUMER_KEY from env: %w", ErrEmptyEnv)
	}

	return e, nil
}
