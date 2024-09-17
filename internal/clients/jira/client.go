package jira

import (
	"fmt"

	jiralib "github.com/andygrunwald/go-jira/v2/cloud"
)

type Config struct {
	JiraInstance string
	Username     string
	APIToken     string
	Project      string
}

type Client struct {
	cfg  Config
	jira *jiralib.Client
}

func NewClient(cfg Config) (*Client, error) {
	tp := jiralib.BasicAuthTransport{
		Username: cfg.Username,
		APIToken: cfg.APIToken,
	}

	client, err := jiralib.NewClient(cfg.JiraInstance, tp.Client())
	if err != nil {
		return nil, fmt.Errorf("creating jira client: %w", err)
	}

	return &Client{
		cfg:  cfg,
		jira: client,
	}, nil
}
