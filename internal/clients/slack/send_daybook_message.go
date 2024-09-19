package slack

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/color"
	slackapi "github.com/zioyero/go-slack"
	"github.com/zioyero/jira-daybot/internal/daybook"
)

func (c *Client) SendDaybookEntry(ctx context.Context, db *daybook.Daybook) error {
	blocks := c.buildDaybookMessage(db)

	for _, channel := range db.User.DaybookChannels {
		headerBlock := slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject("mrkdwn",
				fmt.Sprintf(":thread: <@%s> *Daybook for %s*", db.User.SlackHandle, db.Day.Format("2006-01-02")), false, false,
			),
			nil,
			nil,
		)

		_, ts, err := c.slack.PostMessageContext(ctx, channel, slackapi.MsgOptionBlocks(headerBlock))
		if err != nil {
			return fmt.Errorf("sending slack message: %w", err)
		}

		_, _, err = c.slack.PostMessageContext(ctx, channel, slackapi.MsgOptionBlocks(blocks...), slackapi.MsgOptionTS(ts))
		if err != nil {
			return fmt.Errorf("sending slack message: %w", err)
		}

		time.Sleep(1 * time.Second) // Sleep for 1 second to avoid rate limiting
	}

	color.Green("Sent daybook entry to Slack")

	return nil
}

func (c *Client) SendDaybookDMReminder(ctx context.Context, db *daybook.Daybook) error {
	blocks := c.buildDaybookMessage(db)

	reviewDaybookReminder := slackapi.NewSectionBlock(
		slackapi.NewTextBlockObject("mrkdwn",
			"Hey there! Daybooks will be reported soon, this is how yours will be reported. :smile:", false, false,
		),
		nil,
		nil,
	)
	blocks = append(blocks, reviewDaybookReminder)

	_, _, err := c.slack.PostMessageContext(ctx, db.User.SlackID, slackapi.MsgOptionBlocks(blocks...))
	if err != nil {
		return fmt.Errorf("sending slack message: %w", err)
	}

	color.Green("Sent daybook reminder to Slack")

	return nil
}
