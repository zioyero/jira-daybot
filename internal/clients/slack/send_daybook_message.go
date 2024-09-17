package slack

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	slackapi "github.com/zioyero/go-slack"
	"github.com/zioyero/jira-daybot/internal/daybook"
)

func (c *Client) SendDaybookEntry(ctx context.Context, db *daybook.Daybook) error {
	blocks := c.buildDaybookMessage(db)

	_, _, err := c.slack.PostMessageContext(ctx, devNullChannel, slackapi.MsgOptionBlocks(blocks...))
	if err != nil {
		return fmt.Errorf("sending slack message: %w", err)
	}

	color.Green("Sent daybook entry to Slack")

	return nil
}

func (c *Client) SendDaybookDMReminder(ctx context.Context, db *daybook.Daybook) error {
	blocks := c.buildDaybookMessage(db)

	_, _, err := c.slack.PostMessageContext(ctx, db.User.SlackID, slackapi.MsgOptionBlocks(blocks...))
	if err != nil {
		return fmt.Errorf("sending slack message: %w", err)
	}

	color.Green("Sent daybook reminder to Slack")

	return nil
}
