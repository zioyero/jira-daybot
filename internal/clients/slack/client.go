package slack

import (
	slackapi "github.com/zioyero/go-slack"
)

const devNullChannel = "C07KPQHT7L7"

type Config struct {
	Token          string
	DaybookChannel string
}

type Client struct {
	config *Config
	slack  *slackapi.Client
}

func NewClient(cfg *Config) *Client {
	return &Client{
		slack:  slackapi.New(cfg.Token),
		config: cfg,
	}
}
